provider "aws" {
  region = var.aws_region
}

# Provide an isolated VPC for the TormentNexus backend
resource "aws_vpc" "marketing_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "tormentnexus-vpc"
    Environment = var.environment
  }
}

# Public subnet for the dashboard
resource "aws_subnet" "public_subnet" {
  vpc_id                  = aws_vpc.marketing_vpc.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name        = "tormentnexus-public"
    Environment = var.environment
  }
}

# Provide a security group
resource "aws_security_group" "agent_sg" {
  name        = "tormentnexus-agent-sg"
  description = "Allow TormentNexus dashboard inbound"
  vpc_id      = aws_vpc.marketing_vpc.id

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Dashboard HTTP"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.admin_ip]
    description = "Admin SSH"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# RDS instance to handle the memory vault and deals
resource "aws_db_instance" "marketing_db" {
  identifier           = "tormentnexus-db"
  engine               = "postgres"
  engine_version       = "16"
  instance_class       = "db.t4g.micro"
  allocated_storage    = 20
  db_name              = "marketing_agent"
  username             = var.db_username
  password             = var.db_password
  skip_final_snapshot  = true
  publicly_accessible  = false
  vpc_security_group_ids = [aws_security_group.agent_sg.id]

  tags = {
    Name        = "tormentnexus-postgres"
    Environment = var.environment
  }
}
