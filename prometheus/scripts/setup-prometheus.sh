#!/usr/bin/env bash

PATH=$PATH:/usr/local/bin

source ./prometheus/scripts/environmets.sh > /dev/null 2>&1 || source environmets.sh > /dev/null 2>&1
source ./prometheus/scripts/functions.sh > /dev/null 2>&1 || source functions.sh > /dev/null 2>&1
# Create a grafana API key
CreateGRFAPIKey
if [[ "$GRF_API_KEY" == "null" ]]; then
    # Delete existing key
    GRF_API_KEY_ID=$(curl -s --insecure -XGET $GRF_SERVER_URL/api/auth/keys |jq .[].id)
    curl -s --insecure -XDELETE $GRF_SERVER_URL/api/auth/keys/$GRF_API_KEY_ID >/dev/null
    # and recreate
    CreateGRFAPIKey
fi
echo -n "Create a grafana API key:" && \
echo -ne "\t\t\t" && Done
sleep 1

# Create Prometheus Datasource
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-H "Authorization:Bearer $GRF_API_KEY" \
-X POST --data "$(CreatePrometheusDataSourcePD)" "$GRF_SERVER_URL/api/datasources"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Grafana datasource is already exists" && \
        echo -ne "\t\t\t" && Skip
    else
        echo ""
        echo -n "Create Grafana datasource for Prometheus:"
        echo -ne "\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Create Grafana datasource for Prometheus:"
    echo -ne "\t\t" && Done
    sleep 1
fi

# Import Conversation dashboard
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type: application/json;charset=UTF-8" \
-H "Authorization:Bearer $GRF_API_KEY" \
-d "@./grafana/dashboards/conversation_dashboard.json" \
-X POST "$GRF_SERVER_URL/api/dashboards/db" |jq .)

if [[ "$POST" == *"success"* ]]; then
    echo -n "Import Conversation dashboard:"
    echo -ne "\t\t\t" && Done
    sleep 1
else
    echo ""
    echo -n "Import Conversation dashboard:"
    echo -ne "\t\t" && Failed
    echo -n "An error occured. Please check the error output"
    echo $POST |jq .
    sleep 1
fi

# Import Post dashboard
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type: application/json;charset=UTF-8" \
-H "Authorization:Bearer $GRF_API_KEY" \
-d "@./grafana/dashboards/post_dashboard.json" \
-X POST "$GRF_SERVER_URL/api/dashboards/db" |jq .)

if [[ "$POST" == *"success"* ]]; then
    echo -n "Import Post dashboard:"
    echo -ne "\t\t\t" && Done
    sleep 1
else
    echo ""
    echo -n "Import Post dashboard:"
    echo -ne "\t\t" && Failed
    echo -n "An error occured. Please check the error output"
    echo $POST |jq .
    sleep 1
fi

# Import User dashboard
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type: application/json;charset=UTF-8" \
-H "Authorization:Bearer $GRF_API_KEY" \
-d "@./grafana/dashboards/user_dashboard.json" \
-X POST "$GRF_SERVER_URL/api/dashboards/db" |jq .)

if [[ "$POST" == *"success"* ]]; then
    echo -n "Import User dashboard:"
    echo -ne "\t\t\t" && Done
    sleep 1
else
    echo ""
    echo -n "Import User dashboard:"
    echo -ne "\t\t" && Failed
    echo -n "An error occured. Please check the error output"
    echo $POST |jq .
    sleep 1
fi

# Import Counter dashboard
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type: application/json;charset=UTF-8" \
-H "Authorization:Bearer $GRF_API_KEY" \
-d "@./grafana/dashboards/counter_dashboard.json" \
-X POST "$GRF_SERVER_URL/api/dashboards/db" |jq .)

if [[ "$POST" == *"success"* ]]; then
    echo -n "Import Counter dashboard:"
    echo -ne "\t\t\t" && Done
    sleep 1
else
    echo ""
    echo -n "Import Counter dashboard:"
    echo -ne "\t\t" && Failed
    echo -n "An error occured. Please check the error output"
    echo $POST |jq .
    sleep 1
fi
