#!/bin/sh

# fileupload
curl -i -X POST  -F "data=@test.jpeg" http://localhost:32771/api/v1/putimage/50
