package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a GCP resource (Storage Bucket)
		bucket, err := storage.NewBucket(ctx, "my-bucket", &storage.BucketArgs{
			Website: storage.BucketWebsiteArgs{
				MainPageSuffix: pulumi.String("index.html"),
			},
			UniformBucketLevelAccess: pulumi.Bool(true),
			Location:                 pulumi.String("US"),
		})
		if err != nil {
			return err
		}

		_, err = storage.NewBucketIAMBinding(ctx, "my-bucket-IAMBinding", &storage.BucketIAMBindingArgs{
			Bucket: bucket.Name,
			Role:   pulumi.String("roles/storage.objectViewer"),
			Members: pulumi.StringArray{
				pulumi.String("allUsers"),
			},
		})
		if err != nil {
			return err
		}

		bucketObject, err := storage.NewBucketObject(ctx, "index.html", &storage.BucketObjectArgs{
			Bucket:      bucket.Name,
			ContentType: pulumi.String("text/html"),
			Source:      pulumi.NewFileAsset("static/index.html"),
		})

		bucketEndpoint := pulumi.Sprintf("http://storage.googleapis.com/%s/%s", bucket.Name, bucketObject.Name)
		if err != nil {
			return err
		}

		// Export the DNS name of the bucket
		ctx.Export("bucketName", bucket.Url)
		ctx.Export("indexUrl", bucketEndpoint)
		return nil
	})
}
