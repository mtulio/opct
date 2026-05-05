package baseline

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

type baselineIndexItem struct {
	Date             string                 `json:"date"`
	Name             string                 `json:"name"`
	Path             string                 `json:"path"`
	OpenShiftRelease string                 `json:"openshift_version"`
	Provider         string                 `json:"provider"`
	PlatformType     string                 `json:"platform_type"`
	Status           string                 `json:"status"`
	Size             string                 `json:"size"`
	IsLatest         bool                   `json:"is_latest"`
	Tags             map[string]interface{} `json:"tags"`
}
type baselineIndex struct {
	LastUpdate string                        `json:"date"`
	Status     string                        `json:"status"`
	Results    []*baselineIndexItem          `json:"results"`
	Latest     map[string]*baselineIndexItem `json:"latest"`
}

// CreateBaselineIndex lists objects from S3, extracts metadata,
// and calculates the latest by release and platform type, creating an index.json.
// It uses incremental indexing: loads the existing index and only fetches
// metadata for new objects not already in the index.
func (brs *BaselineConfig) CreateBaselineIndex() error {
	svcS3, _, err := brs.createS3Clients()
	if err != nil {
		return fmt.Errorf("failed to create S3 client and validate bucket: %w", err)
	}

	objects, err := ListObjects(svcS3, brs.bucketRegion, brs.bucketName, "api/v0/result/summary/")
	if err != nil {
		return err
	}

	// Load existing index from S3 for incremental updates.
	existingIndex, err := brs.loadIndexFromS3(svcS3)
	if err != nil {
		log.Warnf("Could not load existing index, performing full reindex: %v", err)
	}

	knownPaths := make(map[string]bool)
	if existingIndex != nil {
		for _, item := range existingIndex.Results {
			knownPaths[item.Path] = true
		}
	}

	index := baselineIndex{
		LastUpdate: time.Now().Format(time.RFC3339),
		Latest:     make(map[string]*baselineIndexItem),
	}

	// Carry over existing results (clear IsLatest — recalculated below).
	if existingIndex != nil {
		for _, item := range existingIndex.Results {
			item.IsLatest = false
			index.Results = append(index.Results, item)
		}
	}

	var newCount int
	for _, obj := range objects {
		objectKey := *obj.Key

		name := objectKey[strings.LastIndex(objectKey, "/")+1:]
		if name == "index.json" || strings.HasSuffix(name, "_latest.json") {
			continue
		}

		if knownPaths[objectKey] {
			continue
		}

		newCount++
		item, err := brs.fetchObjectMetadata(svcS3, objectKey, name, obj)
		if err != nil {
			log.Errorf("failed to process object %s: %v", objectKey, err)
			continue
		}
		index.Results = append(index.Results, item)
	}

	log.Infof("Index update: %d existing, %d new, %d total", len(knownPaths), newCount, len(index.Results))

	// Recalculate latest for all results.
	for _, res := range index.Results {
		res.IsLatest = false
		latestIndexKey := fmt.Sprintf("%s_%s", res.OpenShiftRelease, res.PlatformType)
		existing, ok := index.Latest[latestIndexKey]
		if !ok {
			res.IsLatest = true
			index.Latest[latestIndexKey] = res
		} else if existing.Date < res.Date {
			existing.IsLatest = false
			res.IsLatest = true
			index.Latest[latestIndexKey] = res
		}
	}

	// Copy latest to respective path under /<version>_<platform>_latest.json
	for kLatest, latest := range index.Latest {
		latestObjectKey := fmt.Sprintf("api/v0/result/summary/%s_latest.json", kLatest)
		log.Infof("Creating latest object for %q to %q", kLatest, latestObjectKey)
		_, err := svcS3.CopyObject(&s3.CopyObjectInput{
			Bucket:     aws.String(brs.bucketName),
			CopySource: aws.String(fmt.Sprintf("%v/%v", brs.bucketName, latest.Path)),
			Key:        aws.String(latestObjectKey),
		})
		if err != nil {
			log.Errorf("Couldn't create latest object %s: %v", kLatest, err)
		}
	}

	// Save the new index to the bucket.
	indexJSON, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("unable to save index to json: %w", err)
	}

	// Save the index to the bucket
	_, err = svcS3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(brs.bucketName),
		Key:    aws.String(indexObjectKey),
		Body:   strings.NewReader(string(indexJSON)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload index to bucket: %w", err)
	}

	// Expire cache from cloudfront distribution
	svcCloudfront, err := createCloudFrontClient(brs.bucketRegion)
	if err != nil {
		return fmt.Errorf("failed to create cloudfront client: %w", err)
	}
	invalidationPathsStr := []string{
		"/result/summary/index.json",
		"/result/summary/*_latest.json",
	}
	log.Infof("Creating cache invalidation for %v", strings.Join(invalidationPathsStr, " "))
	var invalidationPaths []*string
	for _, path := range invalidationPathsStr {
		invalidationPaths = append(invalidationPaths, aws.String(path))
	}
	_, err = svcCloudfront.CreateInvalidation(&cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(brs.cloudfrontDistributionID),
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: aws.String(time.Now().Format(time.RFC3339)),
			Paths: &cloudfront.Paths{
				Quantity: aws.Int64(int64(len(invalidationPaths))),
				Items:    invalidationPaths,
			},
		},
	})
	if err != nil {

		log.Warnf("failed to create cache invalidation: %v", err)
		fmt.Printf(`Index updated. Run the following command to invalidate index.cache:
aws cloudfront create-invalidation \
	--distribution-id %s \
	--paths %s`, brs.cloudfrontDistributionID, strings.Join(invalidationPathsStr, " "))
		fmt.Println()
	}
	return nil
}

// loadIndexFromS3 reads the existing index.json from S3 for incremental updates.
func (brs *BaselineConfig) loadIndexFromS3(svc *s3.S3) (*baselineIndex, error) {
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(brs.bucketName),
		Key:    aws.String(indexObjectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get index object: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read index body: %w", err)
	}

	var idx baselineIndex
	if err := json.Unmarshal(body, &idx); err != nil {
		return nil, fmt.Errorf("failed to parse index JSON: %w", err)
	}

	// Filter out any _latest entries that may have polluted a previous index.
	var cleaned []*baselineIndexItem
	for _, item := range idx.Results {
		if !strings.HasSuffix(item.Name, "_latest") {
			cleaned = append(cleaned, item)
		}
	}
	idx.Results = cleaned

	return &idx, nil
}

// fetchObjectMetadata downloads a single S3 object and extracts its index metadata.
func (brs *BaselineConfig) fetchObjectMetadata(svc *s3.S3, objectKey, name string, obj *s3.Object) (*baselineIndexItem, error) {
	objReader, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(brs.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", objectKey, err)
	}
	defer objReader.Body.Close()

	body, err := io.ReadAll(objReader.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data %s: %w", objectKey, err)
	}

	bd := &BaselineData{}
	bd.SetRawData(body)
	tags, err := bd.GetSetupTags()
	if err != nil {
		log.Errorf("failed to deserialize tags/metadata from summary data: %v", err)
	}

	log.Infof("Processing summary object: %s", name)
	log.Debugf("Processing metadata: %v", tags)

	openShiftRelease := strings.Split(name, "_")[0]
	if v, ok := tags["openshiftRelease"]; ok {
		openShiftRelease = v.(string)
	} else {
		log.Warnf("missing openshiftRelease tag in metadata, extracting from name: %v", openShiftRelease)
	}

	platformType := strings.Split(name, "_")[1]
	if v, ok := tags["platformType"]; ok {
		platformType = v.(string)
	} else {
		log.Warnf("missing platformType tag in metadata, extracting from name: %v", platformType)
	}

	executionDate := strings.Split(name, "_")[2]
	if v, ok := tags["executionDate"]; ok {
		executionDate = v.(string)
	} else {
		log.Warnf("missing executionDate tag in metadata, extracting from name: %v", executionDate)
	}

	return &baselineIndexItem{
		Date:             executionDate,
		Name:             strings.Split(name, ".json")[0],
		Path:             objectKey,
		Size:             fmt.Sprintf("%d", *obj.Size),
		OpenShiftRelease: openShiftRelease,
		PlatformType:     platformType,
		Tags:             tags,
	}, nil
}

// ListObjects lists all objects in the bucket, paginating through all results.
func ListObjects(svc *s3.S3, bucketRegion, bucketName, path string) ([]*s3.Object, error) {
	var objects []*s3.Object
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(path),
	}
	err := svc.ListObjectsPages(input, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		objects = append(objects, page.Contents...)
		return true
	})
	if err != nil {
		return nil, err
	}
	return objects, nil
}
