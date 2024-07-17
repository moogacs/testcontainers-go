package couchbase_test

import (
	"context"
	"testing"
	"time"

	"github.com/couchbase/gocb/v2"

	tccouchbase "github.com/testcontainers/testcontainers-go/modules/couchbase"
)

const (
	// dockerImages {
	enterpriseEdition = "couchbase:enterprise-7.6.1"
	communityEdition  = "couchbase:community-7.1.1"
	// }
)

func TestCouchbaseWithCommunityContainer(t *testing.T) {
	ctx := context.Background()

	// withBucket {
	bucketName := "testBucket"
	bucket := tccouchbase.NewBucket(bucketName)

	bucket = bucket.WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(false).
		WithPrimaryIndex(true)

	container, err := tccouchbase.Run(ctx, communityEdition, tccouchbase.WithBuckets(bucket))
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	cluster, err := connectCluster(ctx, container)
	if err != nil {
		t.Fatalf("could not connect couchbase: %s", err)
	}

	testBucketUsage(t, cluster.Bucket(bucketName))
}

func TestCouchbaseWithEnterpriseContainer(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	bucket := tccouchbase.NewBucket(bucketName).
		WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(true).
		WithPrimaryIndex(true)
	container, err := tccouchbase.Run(ctx,
		enterpriseEdition,
		tccouchbase.WithBuckets(bucket),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	cluster, err := connectCluster(ctx, container)
	if err != nil {
		t.Fatalf("could not connect couchbase: %s", err)
	}

	testBucketUsage(t, cluster.Bucket(bucketName))
}

func TestWithCredentials(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	_, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithAdminCredentials("testcontainers", "testcontainers.IS.cool!"),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))
	if err != nil {
		t.Errorf("Expected error to be [%v] , got nil", err)
	}
}

func TestWithCredentials_Password_LessThan_6(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	_, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithAdminCredentials("testcontainers", "12345"),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))

	if err == nil {
		t.Errorf("Expected error to be [%v] , got nil", err)
	}
}

func TestAnalyticsServiceWithCommunityContainer(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	_, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithServiceAnalytics(),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))

	if err == nil {
		t.Errorf("Expected error to be [%v] , got nil", err)
	}
}

func TestEventingServiceWithCommunityContainer(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	_, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithServiceEventing(),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))

	if err == nil {
		t.Errorf("Expected error to be [%v] , got nil", err)
	}
}

func testBucketUsage(t *testing.T, bucket *gocb.Bucket) {
	err := bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		t.Fatalf("could not connect bucket: %s", err)
	}

	key := "foo"
	data := map[string]string{"key": "value"}
	collection := bucket.DefaultCollection()

	_, err = collection.Upsert(key, data, nil)
	if err != nil {
		t.Fatalf("could not upsert data: %s", err)
	}

	result, err := collection.Get(key, nil)
	if err != nil {
		t.Fatalf("could not get data: %s", err)
	}

	var resultData map[string]string
	err = result.Content(&resultData)
	if err != nil {
		t.Fatalf("could not assign content: %s", err)
	}

	if resultData["key"] != "value" {
		t.Errorf("Expected value to be [%s], got %s", "value", resultData["key"])
	}
}

func connectCluster(ctx context.Context, container *tccouchbase.CouchbaseContainer) (*gocb.Cluster, error) {
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return gocb.Connect(connectionString, gocb.ClusterOptions{
		Username: container.Username(),
		Password: container.Password(),
	})
}
