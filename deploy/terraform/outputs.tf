output "db_endpoint" {
  description = "The endpoint of the PostgreSQL RDS instance"
  value       = aws_db_instance.marketing_db.endpoint
}

output "vpc_id" {
  description = "The ID of the generated VPC"
  value       = aws_vpc.marketing_vpc.id
}
