#!/bin/bash

for item in 1 .. 10
do
    http GET 127.0.0.1:8000/verify/ id:=1233 token=vasds Auth-Header-V1:121
    http GET 127.0.0.1:8000/projects/ Auth-Header-V1:3121 title=MyProject
done
