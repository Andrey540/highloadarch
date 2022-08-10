#!/usr/bin/env bash

for (( c= 1; c <= 10; c++ ))
do
   newman run GenerateData.postman_collection.json -n 1000
done