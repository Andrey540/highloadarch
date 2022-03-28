#!/usr/bin/env bash

for (( c= 1; c <= 22; c++ ))
do
   newman run GenerateData.postman_collection.json -n 10000
done