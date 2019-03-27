variable key_pair_name {
  description = "The name of the key-pair to use for setting up SSH for the syslog server."
}

variable syslog_server_instance_name {
  description = "The name of the EC2 instance that is used for the syslog server."
}

resource "aws_instance" "syslog_server" {
  ami                    = "${data.aws_ami.ubuntu.id}"
  instance_type          = "t2.micro"
  user_data              = "${data.template_file.syslog_server_user_data.rendered}"
  vpc_security_group_ids = ["${aws_security_group.syslog_server.id}"]
  key_name               = "${var.key_pair_name}"

  tags {
    Name = "${var.syslog_server_instance_name}"
  }
}

resource "aws_security_group" "syslog_server" {
  name = "${var.syslog_server_instance_name}"

  # Egress: open to the world
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Ingress connection to the syslog server
  ingress {
    from_port = "${var.syslog_server_port}"
    to_port   = "${var.syslog_server_port}"
    protocol  = "tcp"

    # To keep this example simple, we allow incoming HTTP requests from any IP.
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH support
  ingress {
    from_port = "22"
    to_port   = "22"
    protocol  = "tcp"

    # To keep this example simple, we allow incoming HTTP requests from any IP.
    cidr_blocks = ["0.0.0.0/0"]
  }
}

data "template_file" "syslog_server_user_data" {
  template = "${file("${path.module}/syslog_server_user_data.sh")}"

  vars {
    port = "${var.syslog_server_port}"
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "image-type"
    values = ["machine"]
  }

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }
}

output syslog_server_instance_id {
  value = "${aws_instance.syslog_server.id}"
}
