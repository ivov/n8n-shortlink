resource "aws_s3_bucket" "backup" {
  bucket = var.bucket_name
}

resource "aws_s3_bucket_lifecycle_configuration" "backup" {
  bucket = aws_s3_bucket.backup.id

  rule {
    id     = "cleanup"
    status = "Enabled"

    filter {
      prefix = "${var.project_name}-backup/"
    }

    expiration {
      days = 10
    }
  }
}

resource "aws_iam_user" "backup" {
  name = "${var.project_name}-backup-user"
}

resource "aws_iam_access_key" "backup" {
  user = aws_iam_user.backup.name
}

resource "aws_iam_user_policy" "backup" {
  name = "${var.project_name}-backup-policy"
  user = aws_iam_user.backup.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject", 
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.backup.arn,
          "${aws_s3_bucket.backup.arn}/*"
        ]
      }
    ]
  })
}
