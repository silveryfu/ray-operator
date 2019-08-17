import os
from typing import List

import boto3
import botocore
import botocore.client
import botocore.errorfactory

s3 = boto3.client('s3')
s3_res = boto3.resource('s3')


def create_bucket(bucket_name: str, region="us-west-1"):
    if not region[-1].isdigit():
        region = region[:-1]

    print(region)
    s3.create_bucket(
        ACL='private',
        Bucket=bucket_name,
        CreateBucketConfiguration={
            'LocationConstraint': region,
        },
    )


def bucket_exist(bucket: str):
    try:
        s3_res.meta.client.head_bucket(Bucket=bucket)
        return True
    except botocore.client.ClientError:
        return False


def empty_bucket(bucket_name: str):
    bucket = s3_res.Bucket(bucket_name)
    bucket.objects.all().delete()


def delete_bucket(bucket_name: str):
    bucket = s3_res.Bucket(bucket_name)
    bucket.objects.all().delete()
    bucket.delete()


def delete_buckets(match=lambda x: False):
    """Use with caution."""
    buckets = list_buckets()

    for bucket in buckets:
        s3_bucket = s3_res.Bucket(bucket)
        if match(bucket):
            print("match:", bucket)
            input("delete {} ?".format(bucket))
            s3_bucket.objects.all().delete()
            s3_bucket.delete()


def create_dir(bucket_name: str, dir_name: str):
    s3.put_object(
        Bucket=bucket_name,
        Key=(dir_name + "/")
    )


def get_object(bucket: str, prefix: str):
    # Since the key is not known ahead of time, we need to get the
    # object's summary first and then fetch the object
    bucket = s3_res.Bucket(bucket)
    objs = [i for i in bucket.objects.filter(Prefix=prefix)]

    return objs[0].get()["Body"].read()


def get_object_last_mod_time(bucket: str, prefix: str):
    bucket = s3_res.Bucket(bucket)
    objs = [i for i in bucket.objects.filter(Prefix=prefix)]

    return objs[0].last_modified


def create_local_file(path):
    basedir = os.path.dirname(path)
    if not os.path.exists(basedir):
        os.makedirs(basedir)
    open(path, "w").close()


def list_buckets() -> List[str]:
    # Call S3 to list current buckets
    response = s3.list_buckets()

    # Get a list of all bucket names from the response
    buckets = [bucket['Name'] for bucket in response['Buckets']]

    # Print out the bucket list
    # print("Bucket List: %s" % buckets)
    return buckets


def get_addr_from_bucket_key(bucket, key):
    return "s3a://" + bucket + "/" + key


if __name__ == '__main__':
    import pprint as pp
    # pp.pprint(list_buckets())
    # exit()
    delete_buckets(lambda x: x in {'02fc8-state-store',
                                   '103d3-state-store',
                                   '2a80e-state-store',
                                   '302d9-state-store',
                                   '3e0e6-state-store',
                                   '41206-state-store',
                                   '443ab-state-store',
                                   '60ac5-state-store',
                                   '60cae-state-store',
                                   '61621-state-store',
                                   '692cf-state-store',
                                   '79398-state-store',
                                   'a490c-state-store',
                                   'b5ac1-state-store',
                                   'c0b86-state-store',
                                   'e747c-state-store',
                                   })
