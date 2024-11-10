# Provisioning

This guide explains how to provision the requisite infrastructure for `n8n-shortlink`:

- **Server**: CAX11 on Hetzner Cloud. ARM64 with 2 vCPU, 4 GiB RAM, 40 GiB disk, running Ubuntu 22.04, located at `nbg1-dc3` (Nuremberg) data center. Ingress rules configured at network level for TCP ports 22 (SSH with source IP restriction), 80 (HTTP), and 443 (HTTPS)
- **Backup**: AWS S3 bucket for backup storage with 10-day retention policy. Dedicated IAM user with programmatic access via access key pair, constrained by least-privilege policy granting bucket-specific read (`GetObject`, `ListBucket`) and write (`PutObject`) permissions.

## Setup

1. Install Terraform:

```sh
brew install terraform
terraform --version # >= 1.9.8
```

2. Create an SSH key pair:

```sh
ssh-keygen -t ed25519 -C "my@email.com" -f ~/.ssh/id_ed25519_shortlink_via_terraform
```

3. Sign up for [Hetzner Cloud](https://www.hetzner.com/cloud/), create a project `n8n-shortlink`, and create an API token for the project.

4. Sign up for [HCP Terraform](https://www.hashicorp.com/products/terraform), create an organization `ivov` and a workspace `n8n-shortlink`. In workspace settings, set execution mode to `local`, so that apply occurs locally and HCP Terraform is used only to store state. Set these workspace variables:

- `ssh_public_key`: Content of `~/.ssh/id_ed25519_shortlink_via_terraform.pub` from step 2. Mark as sensitive.
- `hcloud_token`: API token from step 3. Mark as sensitive.
- `allowed_ssh_ips`: `["your-ip-address"]`, i.e. string array in CIDR notation. Mark as sensitive and _as HCL-type variable_.

5. Sign up for [AWS](https://aws.amazon.com/console/), create an IAM policy `n8n-shortlink-terraform-automation-policy`, create an IAM user `n8n-shortlink-terraform-automation-user` (disallow AWS Management Console access) attaching the policy to this user, generate access keys for this user (select "Third-party service") and store them in HCP Terraform:

- `tf_automation_aws_access_key_id`: Access key ID for `terraform-automation` IAM user. Mark as sensitive.
- `tf_automation_aws_secret_access_key`. Secret access key for `terraform-automation` IAM user. Mark as sensitive.

```jsonc
// n8n-shortlink-terraform-automation-policy
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "iam:*",
      "Resource": "arn:aws:iam::*:user/n8n-shortlink-terraform-backup-user",
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::n8n-shortlink-terraform-backup-bucket",
        "arn:aws:s3:::n8n-shortlink-terraform-backup-bucket/*",
      ],
    },
  ],
}
```

## Provision

1. Log in to HCP Terraform:

   ```sh
   terraform login
   ```

2. Initialize Terraform:

   ```sh
   cd infra/provision
   terraform init
   ```

3. Plan and apply:

   ```sh
   terraform plan
   terraform apply
   ```

4. Note down IP address and retrieve credentials from state:

   ```sh
    Outputs:
    backup_credentials = (sensitive value)
    server_ip = "87.148.121.19"

    terraform output -json backup_credentials
   ```
