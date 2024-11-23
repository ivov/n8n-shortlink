# Provision

This guide explains how to provision the requisite infrastructure for `n8n-shortlink`

- **Server**: CAX11 on Hetzner Cloud. ARM64 with 2 vCPU, 4 GiB RAM, 40 GiB disk, running Ubuntu 22.04, located at `nbg1-dc3` (Nuremberg) data center. Cloud-level ingress rules allow traffic via TCP ports 22 (with IP restriction), 80 and 443.
- **Object store**: AWS S3 bucket for backup storage with 10-day retention policy. Dedicated IAM user with least-privilege policy granting bucket-specific permissions.

## Setup

1. Install Terraform:

```sh
brew install terraform
terraform --version # >= 1.9.8
```

2. Create an SSH key pair:

```sh
ssh-keygen -t ed25519 -C "my@email.com" -f ~/.ssh/id_ed25519_n8n_shortlink_infra
```

3. At [Hetzner Cloud](https://www.hetzner.com/cloud/):

- Sign up for an account
- Create a project `n8n-shortlink`
- Create an API token for the project

4. At [AWS](https://aws.amazon.com/console/):

- Sign up for an account
- Create an IAM policy `n8n-shortlink-infra-admin-policy` (see content below).
- Create an IAM user `n8n-shortlink-infra-admin-user` (no AWS Management Console access), attaching the policy to this admin user.
- Generate access keys for this admin user, selecting "Third-party service". 

Policy: `n8n-shortlink-infra-admin-policy`

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "iam:*",
      "Resource": "arn:aws:iam::*:user/n8n-shortlink-infra-backup-user"
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::n8n-shortlink-infra-backup-bucket",
        "arn:aws:s3:::n8n-shortlink-infra-backup-bucket/*"
      ]
    }
  ]
}
```

5. At [HCP Terraform](https://www.hashicorp.com/products/terraform):

- Sign up for an account
- Create a new organization `n8n-shortlink-infra` 
- Create a new workspace `n8n-shortlink-infra` ("CLI-driven workflow")
- In workspace settings, set execution mode to `remote`
- At the organization level, create a new variable set `n8n-shortlink-infra-variable-set`, apply it to the `n8n-shortlink-infra` workspace, and add these variables, marking them all as sensitive:

  - `ssh_public_key`: Content of `~/.ssh/id_ed25519_n8n_shortlink_infra.pub` from step 3.
  - `hcloud_token`: Hetzner cloud project API token from step 3.
  - `allowed_ssh_ips`: IP address to SSH from in string array in CIDR notation, e.g. `["125.124.122.231/32"]`. Mark as HCL-type variable.
  - `aws_access_key_id`: Access key ID for `n8n-shortlink-infra-backup-user` IAM user.
  - `aws_secret_access_key`. Secret access key for `n8n-shortlink-infra-backup-user` IAM user.

## Run

1. Log in to HCP Terraform:

   ```sh
   terraform login
   ```

2. Initialize Terraform:

   ```sh
   cd infrastructure/01-provision
   terraform init
   ```

3. Plan and apply:

   ```sh
   terraform plan
   terraform apply

   > Apply complete! Resources: 8 added, 0 changed, 0 destroyed.

   > Outputs:
   > backup = (sensitive value)
   > server_ip = "<redacted>"
   ```

4. Parlay Terraform state into Ansible inputs:

   ```sh
    echo "[server]\n$(terraform output -raw server_ip) ansible_user=root ansible_ssh_private_key_file=~/.ssh/id_ed25519_n8n_shortlink_infra" > ../02-configure/hosts

    terraform output -json backup | jq --indent 2 '.' > ../02-configure/tf-output-aws.json
   ```
