#!/usr/bin/env bash

mockgen -source queue.go  -package queue -destination queue_mock.go
