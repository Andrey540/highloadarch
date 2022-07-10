#!/usr/bin/env bash

PATH=$PATH:/usr/local/bin

source ./zabbix/scripts/environmets.sh > /dev/null 2>&1 || source environmets.sh > /dev/null 2>&1
source ./zabbix/scripts/functions.sh > /dev/null 2>&1 || source functions.sh > /dev/null 2>&1

########## ZABBIX DEPLOYMENT ##########
echo ""
echo -e '\E[1m'"\033\DOCKERIZED ZABBIX DEPLOYMENT AND CONFIGURATION SCRIPT \033[0m"
echo -e '\E[1m'"\033\Version: 3.0.0"
echo ""
echo -e '\E[1m'"\033\By this script, the steps listed below will be done: \033[0m"
echo ""
echo -e '\E[1m'"\033\- Latest Docker(CE) engine and docker-compose installation. \033[0m"
echo -e '\E[1m'"\033\- Dockerized zabbix server deployment by using the official zabbix docker images and compose file. \033[0m"
echo -e '\E[1m'"\033\- Required packages installation like epel-repo and jq.\033[0m"
echo -e '\E[1m'"\033\- Creating auto registration actions for Linux & Windows hosts. \033[0m"
echo -e '\E[1m'"\033\- Creating some additional check items/triggers for Linux & Windows templates. \033[0m"
echo -e '\E[1m'"\033\- Grafana integration and deployment of some useful custom dashboards. \033[0m"
echo -e '\E[1m'"\033\- SMTP settings and admin email configurations. (Optional) \033[0m"
echo -e '\E[1m'"\033\- Slack integration. (Optional) \033[0m"
echo ""
echo -e '\E[1m'"\033\NOTE: Any deployed zabbix server containers will be taken down and re-created.\033[0m"

# Wait until zabbix getting up
GetZabbixAuthToken
echo -e '\E[1m'"\033\- Waiting for Zabbix server getting ready... \033[0m"
while [ "$ZBX_AUTH_TOKEN" == "null" ] || [ -z "$ZBX_AUTH_TOKEN" ]
do
    sleep 2
    GetZabbixAuthToken
    echo -e '\E[1m'"\033\- Waiting for Zabbix server getting ready... \033[0m"
done
echo ""
echo -n "Zabbix deployment:" && \
echo -ne "\t\t\t" && Done
sleep 1
EchoDash

########## HOST GROUPS CONFIGURATIONS ##########
# This creates all defined host groups in environment file
echo -e ""
echo -e '\E[96m'"\033\- Create hosts groups.\033[0m"
sleep 1
CreateHostGroups
sleep 1
EchoDash

########## AUTO REGISTRATION CONFIGURATIONS ##########
# Get Windows host group ID to use it on creating the auto registration action.
WGROUPID=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(WinHostGroupIDPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .result[].groupid |tr -d '"')

# Create an auto registration action for Linux servers
echo -e ""
echo -e '\E[96m'"\033\- Create auto registration actions.\033[0m"
sleep 1
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(AutoRegisterLinuxPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Linux auto registration action is already exists."
        echo -ne "\t\t" && Skip
    else

        echo ""
        echo -n "Linux auto registration action:"
        echo -ne "\t\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Linux auto registration action:"
    echo -ne "\t\t\t\t" && Done
    sleep 1
fi

# Create an auto registration action for Windows servers
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(AutoRegisterWinPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Windows auto registration action is already exists."
        echo -ne "\t\t" && Skip
    else
        echo ""
        echo -n "Win auto registration action:"
        echo -ne "\t\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Win auto registration action:"
    echo -ne "\t\t\t\t" && Done
    sleep 1
    EchoDash
fi


########## TEMPLATE CONFIGURATIONS ##########
echo -e ""
echo -e '\E[96m'"\033\- Tune Linux OS Template.\033[0m"
sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(LLDFSRuleLinuxPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "LLD rule is already set to 1m."
        echo -ne "\t\t" && Skip
    else
        echo ""
        echo -n "Set filesystem discovery LLD interval to 30m:"
        echo -ne "\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Set filesystem discovery LLD interval to 30m:" && \
    echo -ne "\t\t" && Done
    sleep 1
fi

sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(FreeDiskSpacePercentLinuxPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Create an item protoype to get free disk space as percentage for all partitions:"
        echo -ne "\t\t" && Skip
    else
        echo ""
        echo -n "Create an item protoype to get free disk space as percentage for all partitions:"
        echo -ne "\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Create an item protoype to get free disk space as percentage for all partitions:" && \
    echo -ne "\t\t" && Done
    sleep 1
fi

sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(LLDNetIfRuleLinuxPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "LLD rule is already set to 30m."
        echo -ne "\t\t\t\t" && Skip
    else
        echo -n "Set netif discovery LLD interval to 30m:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Set netif discovery LLD interval to 30m:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi
sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(TotalMemoryCheckIntervalPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Total memory check interval is already set to 5m."
        echo -ne "\t\t\t\t" && Skip
    else
        echo -n "Set total memory check interval to 5m:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Set total memory check interval to 5m:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi
sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(TotalSwapCheckIntervalPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Total swap check interval is already set to 30m."
        echo -ne "\t\t\t\t" && Skip
    else
        echo -n "Set total swap check interval to 30m:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Set total swap check interval to 30m:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(NumberofCpusIntervalPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Number of CPUs interval is already set to 5m."
        echo -ne "\t\t\t\t" && Skip
    else
        echo -n "Set Number of CPUs interval to 5m:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Set Number of CPUs interval to 5m:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi
EchoDash

echo -e ""
echo -e '\E[96m'"\033\- Tune Windows OS Template.\033[0m"
sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(FreeDiskSpacePercentWinPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Create an item protoype to get free disk space as percentage for all partitions:"
        echo -ne "\t\t" && Skip
    else
        echo ""
        echo -n "Create an item protoype to get free disk space as percentage for all partitions:"
        echo -ne "\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Create an item protoype to get free disk space as percentage for all partitions:" && \
    echo -ne "\t\t" && Done
    sleep 1
fi

sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(LLDFSRuleWinPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "LLD rule is already set to 30m."
        echo -ne "\t\t" && Skip
    else
        echo ""
        echo -n "Set filesystem discovery LLD interval to 30m:"
        echo -ne "\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Set filesystem discovery LLD interval to 30m:" && \
    echo -ne "\t\t" && Done
    sleep 1
fi
sleep 1

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(LLDNetIfRuleWinPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "LLD rule is already set to 30m."
        echo -ne "\t\t\t\t" && Skip
    else
        echo -n "Set netif discovery LLD interval to 30m:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Set netif discovery LLD interval to 30m:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(FreeMemPercentWinPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Free mem in % item is already exist."
        echo -ne "\t\t\t" && Skip
    else
        echo -n "Create fee mem item as percentage:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Create fee mem item as percentage:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(DisableAnnoyingWinServiceDiscovery)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Annoying Windows service discovery already disabled."
        echo -ne "\t" && Skip
    else
        echo ""
        echo -n "Disable annoying Windows service LLD rule:"
        echo -ne "\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Disable annoying Windows service items LLD rule:"
    echo -ne "\t" && Done
    sleep 1
fi

########## ZABBIX API USER CONFIGURATIONS ##########
echo -e ""
echo -e '\E[96m'"\033\- Create a read-only user for Zabbix API.\033[0m"
sleep 1

# Generate an array variable and fill it with created group IDs for API user read permission
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(HostGroupIDSPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)
GRP_IDS=$(echo $POST |jq .result[].groupid |tr -d '"' |sed ':a;N;$!ba;s/\n/ /g')
unset IFS
GRP_IDS_ARRAY=( $GRP_IDS )

# Create a group for API user
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(CreateAPIUserGroupPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "API user group is already exists"
        echo -ne "\t" && Skip
    else
        echo ""
        echo -n "Create API user group:"
        echo -ne "\t\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Create API user group:"
    echo -ne "\t\t\t\t" && Done
    sleep 1
fi

# Get API User Group ID
API_USERS_GROUP_ID=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(GetAPIUserGroupIDPD)" "$ZBX_SERVER_URL/api_jsonrpc.php" \
|jq '.result[] | select(.name == "API Users") | .usrgrpid' | tr -d '"')

# Create an user for API
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(CreateAPIUserPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "API user is already exists"
        echo -ne "\t\t\t" && Skip
    else
        echo ""
        echo -n "Create API user:"
        echo -ne "\t\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Create API user:"
    echo -ne "\t\t\t\t" && Done
    sleep 1
    EchoDash
fi

########## ZABBIX AGENT CONFIGURATIONS ##########
echo -e ""
echo -e '\E[96m'"\033\- Monitor Zabbix Server itself.\033[0m"
sleep 1

# Get Zabbix server Host ID
GetZabbixAuthToken
ZBX_AGENT_HOST_ID=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(GetHostIDPD)" "$ZBX_SERVER_URL/api_jsonrpc.php" \
|jq '.result[] | select(.name == "Zabbix server") | .hostid' | tr -d '"')

# Change zabbix server's host interface to use DNS instead of IP
# in order to connect dockerized zabbix-agent.
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(UpdateHostInterfacePD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
        echo ""
        echo -n "Update Zabbix host interface:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
else
    echo -n "Update Zabbix host interface:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi

# Get zabbix-agent container ID and enable it to become a monitored host.
ZBX_AGENT_CONTAINER_ID=$(docker ps |egrep zabbix-agent |awk '{print $1}')

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(EnableZbxAgentonServerPD)" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
        echo ""
        echo -n "Enable Zabbix agent:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
else
    echo -n "Enable Zabbix agent:"
    echo -ne "\t\t\t\t" && Done
    sleep 1
    EchoDash
fi

########## GRAFANA CONFIGURATIONS ##########
echo -e ""
echo -e '\E[96m'"\033\- Grafana configurations.\033[0m"
sleep 1
# Wait for grafana server to be ready
for (( i=0; i<23; ++i)); do
    GRAFANA_HEALTH=$(curl -s --insecure http://localhost:3000/healthz)
    [ "$GRAFANA_HEALTH" == "Ok" ] && break
    echo -e '\E[1m'"\033\- Waiting for 5 seconds to grafana server getting be ready... ( $(expr $(echo 23) - $i) retries left ) \033[0m"
    sleep 5
done

if [[ "$GRAFANA_HEALTH" != "Ok" ]]; then
        echo -e "\e[91m- [ERROR]: Grafana server still not ready after 2 minutes. Please check grafana container.\e[0m"
        exit 1
fi
# Enable zabbix datasource plugin
POST=$(curl --insecure -s \
-H "Content-Type:application/x-www-form-urlencoded" \
-X POST $GRF_SERVER_URL/api/plugins/alexanderzobnin-zabbix-app/settings?enabled=true)

if [[ "$POST" == *"error"* ]]; then
        echo ""
        echo -n "Enable grafana zabbix plugin:"
        echo -ne "\t\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
else
    echo -n "Enable grafana zabbix plugin:"
    echo -ne "\t\t\t" && Done
    sleep 1
fi

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

# Create Zabbix Datasource
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-H "Authorization:Bearer $GRF_API_KEY" \
-X POST --data "$(CreateZbxDataSourcePD)" "$GRF_SERVER_URL/api/datasources"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "Grafana datasource is already exists" && \
        echo -ne "\t\t\t" && Skip
    else
        echo ""
        echo -n "Create Grafana datasource for zabbix:"
        echo -ne "\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
    fi
else
    echo -n "Create Grafana datasource for zabbix:"
    echo -ne "\t\t" && Done
    sleep 1
fi

# Get uid of the dashboard
ZABBIX_DASHBOARD_ID=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-H "Authorization:Bearer $GRF_API_KEY" \
-X GET "$GRF_SERVER_URL/api/search?folderIds=0&query=&starred=false" |jq .[].uid |tr -d '"')

# Delete existing Zabbix Server Dashboard
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-H "Authorization:Bearer $GRF_API_KEY" \
-X DELETE "$GRF_SERVER_URL/api/dashboards/uid/$ZABBIX_DASHBOARD_ID"  |jq .)

if [[ "$POST" == *"Not found"* ]]; then
        echo -n "Default zabbix dashboard not found."
        echo -ne "\t\t" && Skip
elif [[ "$POST" == *"error"* ]]; then
        echo ""
        echo -n "Delete default zabbix dashboard:"
        echo -ne "\t\t" && Failed
        echo -n "An error occured. Please check the error output"
        echo $POST |jq .
        sleep 1
else
    echo -n "Delete default zabbix dashboard:"
    echo -ne "\t\t" && Done
    sleep 1
fi

# Import Linux servers dashboard
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type: application/json;charset=UTF-8" \
-H "Authorization:Bearer $GRF_API_KEY" \
-d "@./grafana/dashboards/linux_servers_dashboard.json" \
-X POST "$GRF_SERVER_URL/api/dashboards/db" |jq .)

if [[ "$POST" == *"success"* ]]; then
    echo -n "Import Linux servers dashboard:"
    echo -ne "\t\t\t" && Done
    sleep 1
else
    echo ""
    echo -n "Import Linux servers dashboard:"
    echo -ne "\t\t" && Failed
    echo -n "An error occured. Please check the error output"
    echo $POST |jq .
    sleep 1
fi

# Import Zabbix system status dashboard
POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type: application/json;charset=UTF-8" \
-H "Authorization:Bearer $GRF_API_KEY" \
-d "@./grafana/dashboards/zabbix-system-status.json" \
-X POST "$GRF_SERVER_URL/api/dashboards/db" |jq .)

if [[ "$POST" == *"success"* ]]; then
    echo -n "Import Zabbix system status dashboard:"
    echo -ne "\t\t" && Done
    sleep 1
else
    echo ""
    echo -n "Import Zabbix system status dashboard:"
    echo -ne "\t\t" && Failed
    echo -n "An error occured. Please check the error output"
    echo $POST |jq .
    sleep 1
fi
EchoDash
