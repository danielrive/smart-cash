module "vpc" {
  source                                 = "terraform-aws-modules/vpc/aws"
  version                                = "4.0.1"
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



### Create Private Link to access to ECR

###############################################
# AWS private link endpoint to ECR
# In ECN, it is not working as expected
###############################################

resource "aws_vpc_endpoint" "ecr_dkr_vpc_endpoint" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${var.region}.ecr.dkr"
  auto_accept         = true
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  private_dns_enabled = true
  security_group_ids  = [aws_security_group.allow_tls.id]
}

resource "aws_vpc_endpoint" "ecr_api_vpc_endpoint" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${var.region}.ecr.api"
  auto_accept         = true
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  private_dns_enabled = true
  security_group_ids  = [aws_security_group.allow_tls.id]
}


### AWS VPC S3 GATEWAY ENDPOINT

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = module.vpc.vpc_id
  service_name = "com.amazonaws.${var.region}.s3"
}

resource "aws_vpc_endpoint_route_table_association" "s3_endpoint_association" {
  count           = length(module.vpc.private_route_table_ids)
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
  route_table_id  = module.vpc.private_route_table_ids[count.index]
}