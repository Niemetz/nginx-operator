import boto3

# Create an S3 client
s3 = boto3.client('s3')

# List objects in a given bucket
bucket_name = 'john-01-12-2024'
response = s3.list_objects_v2(Bucket=bucket_name)

for obj in response.get('Contents', []):
    print(obj['Key'])
