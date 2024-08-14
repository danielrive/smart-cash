module "vpc" {
  source                                 = "terraform-aws-modules/vpc/aws"
  version                                = "5.9.0"
  name                                   = "vpc-${var.project_name}-${var.environment}"
  cidr                                   = var.cidr
  azs                                    = var.availability_zones
  private_subnets                        = var.private_subnets
  public_subnets                         = var.public_subnets
  database_subnets                       = var.db_subnets
  create_database_subnet_group           = var.create_db_subnet_group
  create_database_nat_gateway_route      = false
  create_database_internet_gateway_route = false
  enable_nat_gateway                     = var.enable_nat_gw
  single_nat_gateway                     = var.single_nat_gw
  one_nat_gateway_per_az                 = var.one_nat_per_az
  tags                                   = var.tags
  enable_dns_hostnames                   = true
  enable_dns_support                     = true
  map_public_ip_on_launch                = var.enable_auto_public_ip
}


#### Security Group for ecr endpoints

resource "aws_security_group" "allow_tls" {
  name        = "allow_tls"
  description = "Allow TLS inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "TLS from VPC"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.cidr]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = "allow_tls"
  }
}

###############################################
### AWS private link endpoint to ECR

resource "aws_vpc_endpoint" "ecr_dkr_vpc_endpoint" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${var.region}.ecr.dkr"
  auto_accept         = true
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  private_dns_enabled = true
  security_group_ids  = [aws_security_group.allow_tls.id]
  tags = {
    Name = "ecr-endp-${var.project_name}-${var.environment}"
  }
}

# Policy for ECR endpoint

resource "aws_vpc_endpoint_policy" "ecr" {
  vpc_endpoint_id = aws_vpc_endpoint.ecr_dkr_vpc_endpoint.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement": [
      {
        "Sid": "LimitECRAccess",
        "Principal": "*",
        "Action": "*",
        "Effect": "Allow",
        "Resource": "arn:aws:ecr:${var.region}:${var.account_id}:repository/*"
      }
    ]
    })
}

resource "aws_vpc_endpoint" "ecr_api_vpc_endpoint" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${var.region}.ecr.api"
  auto_accept         = true
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  private_dns_enabled = true
  security_group_ids  = [aws_security_group.allow_tls.id]
  tags = {
    Name = "ecr-api-endp-${var.project_name}-${var.environment}"
  }
}

# Policy for ECR endpoint

resource "aws_vpc_endpoint_policy" "ecr-api" {
  vpc_endpoint_id = aws_vpc_endpoint.ecr_api_vpc_endpoint.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement": [
      {
        "Sid": "LimitECRAccess",
        "Principal": "*",
        "Action": "*",
        "Effect": "Allow",
        "Resource": "arn:aws:ecr:${var.region}:${var.account_id}:repository/*"
      }
    ]
    })
}

### AWS VPC S3 GATEWAY ENDPOINT

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = module.vpc.vpc_id
  service_name = "com.amazonaws.${var.region}.s3"
  tags = {
    Name = "s3-endp-${var.project_name}-${var.environment}"
  }
}

resource "aws_vpc_endpoint_route_table_association" "s3_endpoint_association" {
  count           = length(module.vpc.private_route_table_ids)
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
  route_table_id  = module.vpc.private_route_table_ids[count.index]
}

### AWS VPC DynamoDB endpoint
resource "aws_vpc_endpoint" "dynamodb" {
  vpc_id       = module.vpc.vpc_id
  service_name = "com.amazonaws.${var.region}.dynamodb"
  tags = {
    Name = "dynamodb-endp-${var.project_name}-${var.environment}"
  }
}

resource "aws_vpc_endpoint_route_table_association" "dynamodb" {
  count           = length(module.vpc.private_route_table_ids)
  vpc_endpoint_id = aws_vpc_endpoint.dynamodb.id
  route_table_id  = module.vpc.private_route_table_ids[count.index]
}