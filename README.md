modern-app
==========

Starter stack and best practices for building a modern, efficient, scalable,
cheap web app on AWS.

Overview of the stack:

	                            User
	                             |
	                   +---------+---------+
	                   |                   |
	                   |              React router
	frontend           |                   |
	                   |                 React
	                   |                   |
	                   |                Apollo
	                   |                   |
	              *********************************
	              *         the internet          *
	              *********************************
	                   |                   |
	               CloudFront         API Gateway
	                   |                   |
	backend           S3                 Lambda
	                                       |
	                                   Database

instructions
------------

1. `terraform init`
