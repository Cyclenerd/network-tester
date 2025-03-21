# Copyright 2025 Nils Knieling. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM docker.io/library/golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o network-tester .

# Create final lightweight image
FROM docker.io/library/debian:bookworm

WORKDIR /app

# Install network tools
RUN apt-get update -yq && \
	apt-get install -yqq \
		curl \
		dnsutils \
		iperf3 \
		iputils-ping \
		net-tools \
		traceroute && \
	apt-get clean

# Copy the binary from builder
COPY --from=builder /app/network-tester .

# Copy static files
COPY --from=builder /app/static ./static

# Expose the port
EXPOSE 8080

# Run the application
CMD ["./network-tester"]
