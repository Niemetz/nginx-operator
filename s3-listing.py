import boto3

s3 = boto3.client('s3')


def list_s3_bucket(bucket_name):
    response = s3.list_objects_v2(Bucket=bucket_name)
    for obj in response.get('Contents', []):
        print(obj['Key'])


list_s3_bucket('john-01-12-2024')
