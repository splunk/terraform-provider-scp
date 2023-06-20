#!/bin/bash -e

# For more details on VictorOps API, visit: https://help.victorops.com/knowledge-base/rest-endpoint-integration-guide/

help() {
    script_name=$(basename $0)

    echo "This script allows to trigger VictorOps alerts for ACS."
    echo "Usage: $script_name <entity_id> <entity_display_name>"
    echo ""
    echo "Note: \$NAMEASPACE environment variable must be set to identify which namespace to trigger alerts".
}

if [ -z "${NAMESPACE}" ] ; then
    echo "No \$NAMESPACE set, exiting ..."
    help
    exit 1
fi

# Use this field to set the alert name for the incident.
if [ -z "$1" ] ; then
    echo "No entity_id provided, exiting ..."
    help
    exit 1
fi
# Append current date time for uniqueness for the incident
ENTITY_ID="$1"-$(date +%Y-%m-%d-%H-%M)

# Use this field to give human friendly title for the incident
if [ -z "$2" ]; then
    echo "No entity_display_name provided, exiting ..."
    help
    exit 1
fi
ENTITY_DISPLAY_NAME="$2"

PAYLOAD='{ "message_type": "critical", "entity_id": "'$ENTITY_ID'", "entity_display_name": "'$ENTITY_DISPLAY_NAME'" }'
echo $PAYLOAD

curl -X POST -H 'Content-type: application/json' -d "$PAYLOAD" \
    "https://alert.victorops.com/integrations/generic/20131114/alert/3eb00348-6890-4671-982a-e09b335b5e9e/namespace-${NAMESPACE}"