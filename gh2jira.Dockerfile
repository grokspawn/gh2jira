FROM golang:1.24 AS builder
COPY . .
RUN ["make", "build"]

FROM ubuntu:latest
WORKDIR /bin
COPY ./gh2jira /bin/


