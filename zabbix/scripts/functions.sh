#!/usr/bin/env bash

source ./scripts/environmets.sh > /dev/null 2>&1 || source environmets.sh > /dev/null 2>&1

# Global functions
function Done() {
    echo -e '==> \E[32m'"\033\done\033[0m"
}

function Skip() {
    echo -e '==> \E[32m'"\033\skipped\033[0m"
}

function Failed() {
    echo -e '==> \E[91m'"\033\ failed\033[0m"
}

function EchoDash() {
echo "----------------------------------------------------------------"
}

function FinisMessage() {
        echo ""
        echo -e '\E[1m'"\033\Zabbix installation successfuly finished.\033[0m"
        echo "-----------------------------------------------------------------"
        echo ""
        echo -e '\E[1m'"\033\Zabbix UI is accessible at http://ip:8081 \033[0m"
        echo -e '\E[1m'"\033\Username: Admin \033[0m"
        echo -e '\E[1m'"\033\Pasword: zabbix (Don't forget to change it!)\033[0m"
        echo ""
        echo -e '\E[1m'"\033\Grafana UI is accessible at http://ip:3000 \033[0m"
        echo -e '\E[1m'"\033\Username: admin \033[0m"
        echo -e '\E[1m'"\033\Pasword: zabbix (Don't forget to change it too!)\033[0m"
        echo "-----------------------------------------------------------------"
}

function CheckZabbix() {
    cd $BASEDIR
    status=$(docker-compose ps |egrep zabbix-server |egrep " Up " || echo "Not deployed")
    }

function GetZabbixAuthToken() {
    ZBX_AUTH_TOKEN=$(curl --insecure -s \
    -H "Accept: application/json" \
    -H "Content-Type:application/json" \
     -X POST -d \
     '{"jsonrpc":"2.0",
     "method":"user.login",
     "params":
     {"user":"Admin",
     "password":"zabbix"},
     "auth":null,"id":0}' \
     $ZBX_SERVER_URL/api_jsonrpc.php |jq .result |tr -d '"')
    }

# Create a host group for Windows servers
function CreateHostGroups() {

for i in "${HOST_GROUPS[@]}"
do
PD=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "hostgroup.create",
    "params": {
        "name": "$i"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
)

POST=$(curl -s --insecure \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$PD" "$ZBX_SERVER_URL/api_jsonrpc.php"  |jq .)

if [[ "$POST" == *"error"* ]]; then
    if [[ "$POST" == *"already exists"* ]]; then
        echo -n "$i already exists." && \
        echo -ne "\t\t\t" && Skip
    else
        echo -n "An error occured. Please check the error output." && \
        echo $POST |jq .
        echo -ne "\t\t" && Failed
    fi
else
echo -n "$i:" && \
echo -ne "\t\t\t\t\t" && Done
sleep 1
fi
done
}

# Create Auto Registration for Linux hosts
function AutoRegisterLinuxPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "action.create",
    "params": {
        "name": "Linux Servers",
        "eventsource": 2,
        "status": 0,
        "esc_period": "1h",
        "def_shortdata": "Auto registration: {HOST.HOST}",
        "def_longdata": "Host name: {HOST.HOST}\r\nHost IP: {HOST.IP}\r\nAgent port: {HOST.PORT}",
		"r_shortdata": "",
        "r_longdata": "",
		"ack_shortdata": "",
        "ack_longdata": "",
        "filter": {
            "evaltype": 0,
            "formula": "",
            "conditions": [
                {
                    "conditiontype": "24",
                    "operator": "2",
                    "value": "Linux",
                    "formulaid": "A"
                }
            ]
        },
            "operations": [
                {
                    "actionid": "8",
                    "operationtype": "2",
                    "esc_period": "0",
                    "esc_step_from": "1",
                    "esc_step_to": "1",
                    "evaltype": "0",
                    "opconditions": []
                },
            	{
                    "actionid": "8",
                    "operationtype": "4",
                    "esc_period": "0",
                    "esc_step_from": "1",
                    "esc_step_to": "1",
                    "evaltype": "0",
                    "opconditions": [],
                    "opgroup": [
                        {
                            "operationid": "16",
                            "groupid": "2"
                        }
                    ]
                },
                {
                    "actionid": "10",
                    "operationtype": "6",
                    "esc_period": "0",
                    "esc_step_from": "1",
                    "esc_step_to": "1",
                    "evaltype": "0",
                    "opconditions": [],
                    "optemplate": [
                        {
                            "operationid": "19",
                            "templateid": "10001"
                        }
                    ]
                }
            ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Get win host group id
function WinHostGroupIDPD() {
cat <<EOF
{ "jsonrpc": "2.0",
          "method": "hostgroup.get",
          "params": {
            "output": "extend",
            "filter": {
                "name": [
                    "Windows servers"
                ]
            }
          },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Get host group ids
function HostGroupIDSPD() {
cat <<EOF
{ "jsonrpc": "2.0",
          "method": "hostgroup.get",
          "params": {
            "output": "extend",
            "filter": {
                "name": [
                    "${HOST_GROUPS[0]}",
                    "${HOST_GROUPS[1]}",
                    "${HOST_GROUPS[2]}",
                    "${HOST_GROUPS[3]}",
                    "${HOST_GROUPS[4]}",
                    "${HOST_GROUPS[5]}",
                    "${HOST_GROUPS[6]}",
                    "${HOST_GROUPS[7]}",
                    "${HOST_GROUPS[8]}",
                    "${HOST_GROUPS[9]}",
                    "${HOST_GROUPS[10]}",
                    "${HOST_GROUPS[11]}",
                    "${HOST_GROUPS[12]}",
                    "${HOST_GROUPS[13]}",
                    "${HOST_GROUPS[14]}",
                    "${HOST_GROUPS[15]}",
                    "${HOST_GROUPS[16]}",
                    "${HOST_GROUPS[17]}",
                    "${HOST_GROUPS[18]}",
                    "${HOST_GROUPS[19]}",
                    "${HOST_GROUPS[20]}",
                    "${HOST_GROUPS[21]}",
                    "${HOST_GROUPS[22]}",
                    "${HOST_GROUPS[23]}",
                    "${HOST_GROUPS[24]}",
                    "${HOST_GROUPS[25]}",
                    "${HOST_GROUPS[26]}"
                ]
            }
          },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Create Auto Registration for Windows hosts
function AutoRegisterWinPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "action.create",
    "params": {
        "name": "Windows Servers",
        "eventsource": 2,
        "status": 0,
        "esc_period": "1h",
        "def_shortdata": "Auto registration: {HOST.HOST}",
        "def_longdata": "Host name: {HOST.HOST}\r\nHost IP: {HOST.IP}\r\nAgent port: {HOST.PORT}",
		"r_shortdata": "",
        "r_longdata": "",
		"ack_shortdata": "",
        "ack_longdata": "",
        "filter": {
            "evaltype": 0,
            "formula": "",
            "conditions": [
                {
                    "conditiontype": "24",
                    "operator": "2",
                    "value": "Windows",
                    "formulaid": "A"
                }
            ]
        },
            "operations": [
                {
                    "actionid": "8",
                    "operationtype": "2",
                    "esc_period": "0",
                    "esc_step_from": "1",
                    "esc_step_to": "1",
                    "evaltype": "0",
                    "opconditions": []
                },
            	{
                    "actionid": "8",
                    "operationtype": "4",
                    "esc_period": "0",
                    "esc_step_from": "1",
                    "esc_step_to": "1",
                    "evaltype": "0",
                    "opconditions": [],
                    "opgroup": [
                        {
                            "operationid": "16",
                            "groupid": "$WGROUPID"
                        }
                    ]
                },
                {
                    "actionid": "10",
                    "operationtype": "6",
                    "esc_period": "0",
                    "esc_step_from": "1",
                    "esc_step_to": "1",
                    "evaltype": "0",
                    "opconditions": [],
                    "optemplate": [
                        {
                            "operationid": "19",
                            "templateid": "10081"
                        }
                    ]
                }
            ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

########## ITEM AND TRIGGER CONFIGURATIONS ##########

# Set interval to 30m for FS discovery in linux template
function LLDFSRuleLinuxPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "discoveryrule.update",
    "params": {
        "itemid": "29201",
        "delay": "30m"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Set interval to 30m for network interface discovery in linux template
function LLDNetIfRuleLinuxPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "discoveryrule.update",
    "params": {
        "itemid": "29203",
        "delay": "30m"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Set interval to 5m for total memory check in linux template
function TotalMemoryCheckIntervalPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "item.update",
    "params": {
    	"hostid": "10274",
        "itemid": "29104",
        "delay": "5m"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Set interval to 30m for total swap check in linux template
function TotalSwapCheckIntervalPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "item.update",
    "params": {
    	"hostid": "10274",
        "itemid": "29106",
        "delay": "30m"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Add free disk space as percent for Linux hosts
function FreeDiskSpacePercentLinuxPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "itemprototype.create",
    "params": {
        "name": "{#FSNAME}: Free Disk Space in %",
        "key_": "vfs.fs.size[{#FSNAME},pfree]",
		"value_type": "0",
        "units": "%",
        "hostid": "10273",
        "ruleid": "29201",
        "type": 0,
        "delay": "1m",
        "history": "1w",
        "trends": "365d",
        "status": "0",
        "interfaceid": "0",
        "applications": ["1143"]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Set interval to 5m for Number of Cpus item for linux template (also we're going to remove its current preprocessing
# in order to get this value from Grafana
function NumberofCpusIntervalPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "item.update",
    "params": {
    	"hostid": "10272",
        "itemid": "29087",
        "delay": "5m",
        "preprocessing": [
        ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

function FreeDiskSpacePercentWinPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "itemprototype.create",
    "params": {
        "name": "{#FSNAME}: Free space in %",
        "key_": "vfs.fs.size[{#FSNAME},pfree]",
		"value_type": "0",
        "units": "%",
        "hostid": "10288",
        "ruleid": "29506",
        "type": 0,
        "delay": "1m",
        "history": "1w",
        "trends": "365d",
        "status": "0",
        "interfaceid": "0",
        "applications": ["1204"],
        "application_prototypes": [ "Filesystem {#FSNAME}" ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Set interval to 30m for FS discovery for Windows template
function LLDFSRuleWinPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "discoveryrule.update",
    "params": {
        "itemid": "29509",
        "delay": "30m"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Set interval to 30m for network interface discovery for Windows template
function LLDNetIfRuleWinPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "item.update",
    "params": {
    	"hostid": "10081",
        "itemid": "31426",
        "delay": "30m"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Add free memory usege as percent for Windows
function FreeMemPercentWinPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "item.create",
    "params": {
   	    "hostid": "10287",
        "name": "Free memory in %",
        "key_": "vm.memory.size[pavailable]",
		"value_type": 0,
		"type": 0,
        "units": "%",
        "delay": "1m",
        "status": 0,
        "applications": ["1203"]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Disable  annoying service discovery rules which creates too mant false positive alerts.
function DisableAnnoyingWinServiceDiscovery() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "discoveryrule.update",
    "params": {
		"hostid": "10081",
        "itemid": "30424",
        "status": "1"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

function SMTPConfigPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "mediatype.update",
    "params": {
        "mediatypeid": "1",
        "status": 0,
        "smtp_server": "$SMTPServer",
        "smtp_port": "$SMTPServerPort",
        "smtp_helo": "$SMTPHello",
        "smtp_email": "$SMTPEmail",
        "smtp_security": $SecureConnection,
        "smtp_authentication": $Authentication,
        "username": "$SMTPUsername",
        "passwd": "$SMTPPassword"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Slack related functions
function GetSlackNotifAnswer(){
    while true
    do
        echo -e '\E[96m'"\033\ Do you want to enable slack notifications ? (Yes or No): \033[0m \c"
        read  SlackEnable
        case $SlackEnable in
        Yes|yes|YES|YeS|yeS|yEs) break ;;
        No|no|NO|nO) break ;;
        *) echo -e '\E[91m'"\033\ Please type Yes or No \033[0m"
        esac
    done
    if [[ "$SlackEnable" =~ $yesPattern ]]; then
        echo -e ""
        echo -e '\E[96m'"\033\- Slack settings. \033[0m"
        echo -e '\E[1m'"\033\This section, enables slack notification. \033[0m"

        echo -e '\E[1m'"\033\An slack app must be created within your Slack.com workspace \033[0m"
        echo -e '\E[1m'"\033\as explained at http://git.zabbix.com/projects/ZBX/repos/zabbix/browse/templates/media/slack \033[0m"
        echo -e '\E[1m'"\033\Please create the slack app now and provide its Bot User OAuth Access Token, and slack channel name.\033[0m"
        echo ""
        sleep 1

        # Get Bot Token
        echo -e '\E[96m'"\033\ Enter your Bot User OAuth Access Token: \033[0m \c"
        read SlackBotToken
        while [[ -z $SlackBotToken ]]
        do
          echo -e '\E[91m'"\033\ Please enter your Bot User OAuth Access Token:\033[0m \c"
          read SlackBotToken
        done

        # Get slack channel name to send notifications
        echo -e '\E[96m'"\033\ Enter your slack channel: \033[0m \c"
        read SlackChannel
        while [[ -z $SlackChannel ]]
        do
          echo -e '\E[91m'"\033\ Please enter your channel:\033[0m \c"
          read SlackChannel
        done
        SlackChannel="#$SlackChannel"
    else
        echo -n "Slack notification configuration:" && \
        echo -ne "\t\t" && Skip
        sleep 1
    fi
}

# Globl Macro for ZABBIX.URL
function ZabbixUrlGlobalMacroPD(){
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "usermacro.createglobal",
    "params":  {
        "macro": "{\$ZABBIX.URL}",
        "value": "http://$ZBX_PUBLIC_IP:8081/"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 1
}
EOF
}

# Slack bot token
function SetSlackBotTokenPD(){
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "mediatype.update",
    "params": {
        "mediatypeid": "9",
        "status": 0,
        "parameters": [
                {
                    "name": "zabbix_url",
                    "value": "http://$ZBX_PUBLIC_IP:8081"
                },
                {
                    "name": "bot_token",
                    "value": "$SlackBotToken"
                },
                {
                    "name": "channel",
                    "value": "{ALERT.SENDTO}"
                },
                {
                    "name": "slack_mode",
                    "value": "alarm"
                },
                {
                    "name": "slack_as_user",
                    "value": "true"
                },
                {
                    "name": "event_tags",
                    "value": "{EVENT.TAGS}"
                },
                {
                    "name": "event_nseverity",
                    "value": "{EVENT.NSEVERITY}"
                },
                {
                    "name": "event_value",
                    "value": "{EVENT.VALUE}"
                },
                {
                    "name": "event_update_status",
                    "value": "{EVENT.UPDATE.STATUS}"
                },
                {
                    "name": "event_date",
                    "value": "{EVENT.DATE}"
                },
                {
                    "name": "event_time",
                    "value": "{EVENT.TIME}"
                },
                {
                    "name": "event_severity",
                    "value": "{EVENT.SEVERITY}"
                },
                {
                    "name": "event_opdata",
                    "value": "{EVENT.OPDATA}"
                },
                {
                    "name": "event_id",
                    "value": "{EVENT.ID}"
                },
                {
                    "name": "trigger_id",
                    "value": "{TRIGGER.ID}"
                },
                {
                    "name": "trigger_description",
                    "value": "{TRIGGER.DESCRIPTION}"
                },
                {
                    "name": "host_name",
                    "value": "{HOST.HOST}"
                },
                {
                    "name": "event_update_date",
                    "value": "{EVENT.UPDATE.DATE}"
                },
                {
                    "name": "event_update_time",
                    "value": "{EVENT.UPDATE.TIME}"
                },
                {
                    "name": "event_recovery_date",
                    "value": "{EVENT.RECOVERY.DATE}"
                },
                {
                    "name": "event_recovery_time",
                    "value": "{EVENT.RECOVERY.TIME}"
                },
                {
                    "name": "alert_message",
                    "value": "{ALERT.MESSAGE}"
                },
                {
                    "name": "alert_subject",
                    "value": "{ALERT.SUBJECT}"
                },
                {
                    "name": "discovery_host_dns",
                    "value": "{DISCOVERY.DEVICE.DNS}"
                },
                {
                    "name": "discovery_host_ip",
                    "value": "{DISCOVERY.DEVICE.IPADDRESS}"
                },
                {
                    "name": "event_source",
                    "value": "{EVENT.SOURCE}"
                },
                {
                    "name": "host_conn",
                    "value": "{HOST.CONN}"
                }
            ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 1
}
EOF
}

# User media type configuration
function AdminSmtpMediaTypePD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "user.update",
    "params": {
        "userid": "1",
        "user_medias": [
            {
                "mediatypeid": "1",
                "sendto": "$SentTo",
                "active": 0,
                "severity": 63,
                "period": "1-7,00:00-24:00"
            }
        ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

function AdminSlackMediaTypePD(){
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "user.update",
    "params": {
        "userid": "1",
        "user_medias": [
            {
                "mediatypeid": "9",
                "sendto": "$SlackChannel",
                "active": 0,
                "severity": 63,
                "period": "1-7,00:00-24:00"
            }
        ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

function AdminSmtpSlackMediaTypePD(){
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "user.update",
    "params": {
        "userid": "1",
        "user_medias": [
            {
                "mediatypeid": "9",
                "sendto": "$SlackChannel",
                "active": 0,
                "severity": 63,
                "period": "1-7,00:00-24:00"
            },
            {
                "mediatypeid": "1",
                "sendto": "$SentTo",
                "active": 0,
                "severity": 63,
                "period": "1-7,00:00-24:00"
            }
        ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Notification trigger action for administrators
function NotifTriggerPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "action.update",
    "params": {
        "actionid": 3,
        "status": 0
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# API related
function GetAPIUserGroupIDPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "usergroup.get",
    "params": {
        "output": "extend",
        "status": 0
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 1
}
EOF
}

# Add user group for zabbix api user
function CreateAPIUserGroupPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "usergroup.create",
    "params": {
        "name": "API Users",
        "gui_access": 3,
		"users_status": 0,
		"rights": [
			{
				"permission": 2,
		    	"id": "2"
		    },
		    {
				"permission": 2,
		    	"id": "4"
		    },
			{
				"permission": 2,
		    	"id": "5"
		    },
			{
				"permission": 2,
		    	"id": "6"
		    },
			{
				"permission": 2,
		    	"id": "7"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[0]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[1]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[2]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[3]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[4]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[5]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[6]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[7]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[8]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[9]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[10]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[11]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[12]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[13]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[14]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[15]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[16]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[17]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[18]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[19]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[20]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[21]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[22]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[23]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[24]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[25]}"
		    },
		    {
		    	"permission": 2,
		    	"id": "${GRP_IDS_ARRAY[26]}"
		    }
		]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

function CreateAPIUserPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "user.create",
    "params": {
        "alias": "apiuser",
        "passwd": "zabbix",
        "roleid": "1",
        "usrgrps": [
            {
                "usrgrpid": "$API_USERS_GROUP_ID"
            }
        ]
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

# Zabbix server Host ID
function GetHostIDPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "host.get",
    "params": {
        "output": "extend"
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

function UpdateHostInterfacePD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "hostinterface.update",
    "params": {
        "interfaceid": "1",
        "hostids": "$ZBX_AGENT_HOST_ID",
        "type": 1,
        "useip": 0,
        "dns": "zabbix-agent",
        "port": 10050,
        "main": 1
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

function EnableZbxAgentonServerPD() {
cat <<EOF
{
    "jsonrpc": "2.0",
    "method": "host.update",
    "params": {
        "hostid": "$ZBX_AGENT_HOST_ID",
        "host": "$ZBX_AGENT_CONTAINER_ID",
        "name": "Zabbix server",
        "status": 0
    },
    "auth": "$ZBX_AUTH_TOKEN",
    "id": 0
}
EOF
}

### Grafana related
function CreateGRFAPIKey() {
    GRF_API_KEY=$(curl --insecure -s \
    -H "Accept: application/json" \
    -H "Content-Type:application/json" \
    -X POST -d \
     '{
	    "name":"zabbix-api-key",
	    "role": "Admin"
      }' \
     $GRF_SERVER_URL/api/auth/keys |jq .key |tr -d '"')
}

function CreateZbxDataSourcePD() {
cat <<EOF
{
        "orgId": 1,
        "name": "zabbix",
        "type": "alexanderzobnin-zabbix-datasource",
        "typeLogoUrl": "public/plugins/alexanderzobnin-zabbix-datasource/img/zabbix_app_logo.svg",
        "access": "proxy",
        "url": "http://zabbix-web-nginx-mysql:8080/api_jsonrpc.php",
        "password": "zabbix",
        "user": "apiuser",
        "database": "",
        "basicAuth": false,
        "isDefault": true,
        "jsonData": {
            "dbConnection": {
                "enable": false
            },
            "keepCookies": [],
            "password": "zabbix",
            "tlsSkipVerify": true,
            "username": "apiuser",
            "cacheTTL": "5m"
        },
        "readOnly": false
}
EOF
}

