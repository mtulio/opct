package cleaner

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"regexp"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	log "github.com/sirupsen/logrus"
	"k8s.io/utils/ptr"
)

type PatchRule struct {
	JSONPatch    *string
	RegexPattern *regexp.Regexp
	KeepCount    uint64
	Count        uint64
}

var (
	// JSONPatchRules is a map with the paths to files in the archive to apply RFC6902 JSON patches.
	JSONPatchRules = map[string]*PatchRule{
		"resources/cluster/machineconfiguration.openshift.io_v1_controllerconfigs.json": &PatchRule{
			JSONPatch: ptr.To[string](`[
					{
						"op": "replace",
						"path": "/items/0/spec/internalRegistryPullSecret",
						"value": "REDACTED"
					}
				]`,
			),
		},
	}

	// RemoveFilePatternRules is a map with regular expressions to remove files in the result archive.
	RemoveFilePatternRules = map[string]*PatchRule{
		"packages.operators.coreos.com_v1_packagemanifests.json": &PatchRule{
			RegexPattern: regexp.MustCompile("resources/ns/.*/packages.operators.coreos.com_v1_packagemanifests.json"),
			// Keeping at least one object for auditing.
			KeepCount: 1,
			Count:     0,
		},
	}
)

// ScanPatchTarGzipReaderFor scans and patches the artifact stream, returning the cleaned artifact.
func ScanPatchTarGzipReaderFor(r io.Reader) (resp io.Reader, size int, err error) {
	log.Debug("Scanning the artifact for patches...")
	size = 0

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, size, fmt.Errorf("unable to open gzip file: %w", err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Create a buffer to store the updated tar.gz content
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	// Process the tar headers
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, size, fmt.Errorf("unable to process file in archive: %w", err)
		}

		if err := processTarHeader(header, tarReader, tarWriter); err != nil {
			return nil, size, err
		}
	}

	// Close the writers
	if err := tarWriter.Close(); err != nil {
		return nil, size, fmt.Errorf("closing tarball: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, size, fmt.Errorf("closing gzip: %w", err)
	}

	// Return the updated tar.gz content as an io.Reader
	size = len(buf.Bytes())
	return bytes.NewReader(buf.Bytes()), size, nil
}

// processTarHeader processes the tar header and applies patches or removes files as needed.
func processTarHeader(header *tar.Header, tarReader *tar.Reader, tarWriter *tar.Writer) error {
	// Processing pre-defined patches, including recursively archives inside base.
	if _, ok := JSONPatchRules[header.Name]; ok {
		// Once the pre-defined/hardcoded patch matches with file stream, apply
		// the path according to the extension. Currently only JSON patches
		// are supported.
		log.Debugf("Patch pattern matched for: %s", header.Name)
		if strings.HasSuffix(header.Name, ".json") {
			var patchedFile []byte
			desiredFile, err := io.ReadAll(tarReader)
			if err != nil {
				log.Errorf("Unable to read file in archive: %v", err)
				return fmt.Errorf("unable to read file in archive: %w", err)
			}
			// Apply JSON patch to the file
			patchedFile, err = applyJSONPatch(header.Name, desiredFile)
			if err != nil {
				log.Errorf("Unable to apply patch to file %s: %v", header.Name, err)
				return fmt.Errorf("unable to apply patch to file %s: %w", header.Name, err)
			}

			// Update the file size in the header
			header.Size = int64(len(patchedFile))
			log.Debugf("File %s size %d bytes", header.Name, header.Size)

			// Write the updated file to stream.
			if err := tarWriter.WriteHeader(header); err != nil {
				log.Errorf("Unable to write file header to new archive: %v", err)
				return fmt.Errorf("unable to write file header to new archive: %w", err)
			}
			if _, err := tarWriter.Write(patchedFile); err != nil {
				log.Errorf("Unable to write file data to new archive: %v", err)
				return fmt.Errorf("unable to write file data to new archive: %w", err)
			}
		} else {
			log.Debugf("Unknown extension, skipping patch for file %s", header.Name)
		}

	} else if strings.HasSuffix(header.Name, ".tar.gz") {
		// Recursively scan for .tar.gz files rewriting it back to the original archive.
		// By default sonobuoy writes a base archive, and the result(s) will be inside of
		// the base. So it is required to recursively scan archives to find required, hardcoded,
		// patches.
		log.Debugf("Scanning tarball archive: %s", header.Name)
		resp, size, err := ScanPatchTarGzipReaderFor(tarReader)
		if err != nil {
			return fmt.Errorf("unable to apply patch to file %s: %w", header.Name, err)
		}

		// Update the file size in the header
		header.Size = int64(size)

		// Write the updated header and file content
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, resp)
		if err != nil {
			return err
		}

		// Write archive back to stream.
		if err := tarWriter.WriteHeader(header); err != nil {
			log.Errorf("Unable to write file header to new archive: %v", err)
			return fmt.Errorf("unable to write file header to new archive: %w", err)
		}
		if _, err := tarWriter.Write(buf.Bytes()); err != nil {
			log.Errorf("Unable to write file data to new archive: %v", err)
			return fmt.Errorf("unable to write file data to new archive: %w", err)
		}
	} else {
		skip := false
		for _, rule := range RemoveFilePatternRules {
			if rule.RegexPattern.MatchString(header.Name) {
				if rule.Count >= rule.KeepCount {
					log.Debugf("Skipping file %s due to matching pattern rules", header.Name)
					skip = true
					continue
				}
				rule.Count += 1
			}
		}
		if skip {
			return nil
		}

		// Do nothing: copy unmatched files as-is.
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("error streaming file header to new archive: %w", err)
		}
		if _, err := io.Copy(tarWriter, tarReader); err != nil {
			return fmt.Errorf("error streaming file data to new archive: %w", err)
		}
	}
	return nil
}

// applyJSONPatch applies hardcoded patches to the stream, returning the cleaned file.
func applyJSONPatch(filepath string, data []byte) ([]byte, error) {
	if _, ok := JSONPatchRules[filepath]; !ok {
		return nil, fmt.Errorf("no patch rule for file: %s", filepath)
	}
	patch, err := jsonpatch.DecodePatch([]byte(*JSONPatchRules[filepath].JSONPatch))
	if err != nil {
		return nil, fmt.Errorf("decoding patch: %w", err)
	}

	modified, err := patch.Apply(data)
	if err != nil {
		return nil, fmt.Errorf("applying patch: %w", err)
	}

	return modified, nil
}
