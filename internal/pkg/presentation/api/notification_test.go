package api

import (
	"encoding/json"
	"testing"

	"github.com/matryer/is"
)

func TestUnmarshalLifebuoyNotification(t *testing.T) {
	is := testSetup(t)

	n := Notification{}
	json.Unmarshal([]byte(lifebuoy_notification), &n)

	is.Equal("2022-06-02T08:34:05.237466Z", n.NotifiedAt)
	is.Equal("Lifebuoy", n.Data[0].Type)
	is.Equal("urn:ngsi-ld:Lifebuoy:mybuoy", n.Data[0].Id)
	is.Equal("off", n.Data[0].Status.Value)
	is.Equal(nil, n.Data[0].DeviceState)
}

func TestUnmarshalDeviceNotification(t *testing.T) {
	is := testSetup(t)

	n := Notification{}
	json.Unmarshal([]byte(device_notification), &n)

	is.Equal("2022-06-02T08:34:05.237466Z", n.NotifiedAt)
	is.Equal("Device", n.Data[0].Type)
	is.Equal("urn:ngsi-ld:Device:device-9845A", n.Data[0].Id)
	is.Equal("ok", n.Data[0].DeviceState.Value)
	is.Equal(nil, n.Data[0].Status)
}

func testSetup(t *testing.T) *is.I {
	return is.New(t)
}

const lifebuoy_notification string = `
{
	"id": "urn:ngsi-ld:Notification:419ef219-06f9-40cb-95eb-97d877036dcf",
	"type": "Notification",
	"subscriptionId": "notimplemented",
	"notifiedAt": "2022-06-02T08:34:05.237466Z",
	"data": [
	 {
	  "@context": [
	   "https://schema.lab.fiware.org/ld/context",
	   "https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld"
	  ],
	  "id": "urn:ngsi-ld:Lifebuoy:mybuoy",
	  "status": {
	   "type": "Property",
	   "value": "off"
	  },
	  "type": "Lifebuoy"
	 }
	]
   }
`
const device_notification string = `
{
	"id": "urn:ngsi-ld:Notification:419ef219-06f9-40cb-95eb-97d877036dcf",
	"type": "Notification",
	"subscriptionId": "notimplemented",
	"notifiedAt": "2022-06-02T08:34:05.237466Z",
	"data": [
		{
			"id": "urn:ngsi-ld:Device:device-9845A",
			"type": "Device",
			"category": {
				"type": "Property",
				"value": ["sensor"]
			},
			"batteryLevel": {
				"type": "Property",
				"value": 0.75
			},
			"dateFirstUsed": {
				"type": "Property",
				"value": {
					"@type": "DateTime",
					"@value": "2014-09-11T11:00:00Z"
				}
			},
			"controlledAsset": {
				"type": "Relationship",
				"object": ["urn:ngsi-ld::wastecontainer-Osuna-100"]
			},
			"serialNumber": {
				"type": "Property",
				"value": "9845A"
			},
			"mcc": {
				"type": "Property",
				"value": "214"
			},
			"value": {
				"type": "Property",
				"value": "l%3D0.22%3Bt%3D21.2"
			},
			"refDeviceModel": {
				"type": "Relationship",
				"object": "urn:ngsi-ld:DeviceModel:myDevice-wastecontainer-sensor-345"
			},
			"rssi": {
				"type": "Property",
				"value": 0.86
			},
			"controlledProperty": {
				"type": "Property",
				"value": ["fillingLevel", "temperature"]
			},
			"owner": {
				"type": "Property",
				"value": ["http://person.org/leon"]
			},
			"mnc": {
				"type": "Property",
				"value": "07"
			},
			"ipAddress": {
				"type": "Property",
				"value": ["192.14.56.78"]
			},
			"deviceState": {
				"type": "Property",
				"value": "ok"
			},
			"@context": [
				"https://schema.lab.fiware.org/ld/context",
				"https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld"
			]
		}
	]
   }
`
