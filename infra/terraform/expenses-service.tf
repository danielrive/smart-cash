################################################
########## Resources for expenses-service

#######################
#### DynamoDB tables

### expenses Table

resource "aws_dynamodb_table" "expenses_table" {
  name         = "expenses-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "expenseId"
  range_key    = "date"

  attribute {
    name = "expenseId"
    type = "S"
  }

  attribute {
    name = "date"
    type = "S"
  }

  attribute {
    name = "tag"
    type = "S"
  }

  attribute {
    name = "userId"
    type = "S"
  }

  global_secondary_index {
    name               = "by_userId"
    hash_key           = "userId"
    projection_type    = "INCLUDE"
    non_key_attributes = ["expenseId","currency","date","amount"]
  }

  global_secondary_index {
    name               = "by_tag"
    hash_key           = "userId"
    range_key           = "tag"
    projection_type    = "ALL"
  }
  tags = {
    Name = "expenses_${var.environment}"
  }
  
}


##############################
###### IAM Role K8 SA

resource "aws_iam_role" "expenses-service-role" {
  name = "role-expenses-service-${var.environment}"
  path = "/"
  assume_role_policy = jsonencode({
  Version="2012-10-17"
  Statement =  [
    {
      Effect= "Allow"
      Principal= {
        Federated= "arn:aws:iam::${data.aws_caller_identity.id_account.id}:oidc-provider/${module.eks_cluster.cluster_oidc}"
      },
      Action= "sts:AssumeRoleWithWebIdentity",
      Condition={
        StringEquals= {
          "${module.eks_cluster.cluster_oidc}:aud": "sts.amazonaws.com",
          "${module.eks_cluster.cluster_oidc}:sub": "system:serviceaccount:${var.environment}:sa-expenses-service"
        }
      }
    }
  ]
})
}

####### IAM policy for SA expenses-service

resource "aws_iam_policy" "dynamodb-expenses-service-policy" {
  name        = "policy-dynamodb-expenses-service-${var.environment}"
  path        = "/"
  description = "policy for k8 service account"

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
                "dynamodb:ConditionCheckItem",
                "dynamodb:PutItem",
                "dynamodb:DescribeTable",
                "dynamodb:DeleteItem",
                "dynamodb:GetItem",
                "dynamodb:Query",
                "dynamodb:UpdateItem"
        ]
        Effect   = "Allow"
        Resource = aws_dynamodb_table.expenses_table.arn
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attachment-expenses-policy-role1" {
  policy_arn = aws_iam_policy.dynamodb-expenses-service-policy.arn
  role       = aws_iam_role.expenses-service-role.name
}



#############################
##### ECR Repo

module "ecr_registry_expenses_service" {
  source       = "./modules/ecr"
  name         = "expenses-service"
  project_name = var.project_name
  environment  = var.environment
}


###########################
##### K8 Manifests 

###########################
##### General resources

resource "github_repository_file" "common_resources" {
  depends_on          = [module.eks_cluster,null_resource.bootstrap-flux]
  for_each            = fileset(local.path_tf_repo_flux_common, "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "common/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_common}/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION = var.region
      ENVIRONMENT = var.environment
      ## Variables cert manager
      ARN_CERT_MANAGER_ROLE = "arn:aws:iam::12345678910:role/cert-manager-us-west-2"
      ## Variables for Grafana
      ## Variables for ingress
      
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


###########################
#####


###########################
#####


###########################
#####


###########################
#####