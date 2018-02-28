resource "aws_db_subnet_group" "default" {
  name       = "main"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}

resource "aws_security_group" "access-database" {
  name        = "access-database"
  description = "Security group to allow access to database"
  vpc_id      = "${aws_vpc.default.id}"
}

# TODO more logs (error, general, slow query, and audit)
resource "aws_security_group" "database" {
  name        = "database"
  description = "Security group for database"
  vpc_id      = "vpc-96d67df3"

  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]

    #security_groups = ["${aws_security_group.access-database.id}"]
  }
}

resource "aws_db_instance" "default" {
  identifier                          = "${var.domain}"
  name                                = "${var.domain}"
  engine                              = "mysql"
  allow_major_version_upgrade         = true
  apply_immediately                   = true
  allocated_storage                   = 8
  db_subnet_group_name                = "${aws_db_subnet_group.default.name}"
  iam_database_authentication_enabled = true
  instance_class                      = "db.t2.micro"
  username                            = "master"
  password                            = "${var.db_master_password}"
  vpc_security_group_ids              = ["${aws_security_group.database.id}"]
  skip_final_snapshot                 = true
}

#resource "aws_rds_cluster" "default" {
#  cluster_identifier                  = "seohub"
#  engine                              = "aurora-mysql"
#  master_username                     = "master"
#  master_password                     = "${var.db_master_password}"
#  db_subnet_group_name                = "${aws_db_subnet_group.default.name}"
#  skip_final_snapshot                 = true                                  # TODO remove this for production
#  apply_immediately                   = true                                  # TODO remove this for production
#  vpc_security_group_ids              = ["${aws_security_group.database.id}"]
#  iam_database_authentication_enabled = true
#}
#
#resource "aws_rds_cluster_instance" "default" {
#  count                = 1
#  identifier           = "seohub-${count.index}"
#  engine               = "aurora-mysql"
#  cluster_identifier   = "${aws_rds_cluster.default.id}"
#  instance_class       = "db.t2.micro"
#  db_subnet_group_name = "${aws_db_subnet_group.default.name}"
#  apply_immediately    = true                                  # TODO remove this for production
#}

