package nx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sender := New(json.RawMessage("{\"server\": \"server1\", \"username\": \"user1\", \"password\": \"pass1\"}"))

	assert.Equal(t, "server1", sender.Server)
	assert.Equal(t, "user1", sender.Username)
	assert.Equal(t, "pass1", sender.Password)
}

func TestTrigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/createEvent?caption=&metadata=%7B%22cameraRefs%22%3A%5B%22camera-uuid1%22%5D%7D&source=stampzilla&state=Active", req.URL.String())
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"server\": \"" + server.URL + "\", \"username\": \"user1\", \"password\": \"pass1\"}"))
	err := sender.Trigger([]string{"camera-uuid1"}, "")

	assert.NoError(t, err)
}

func TestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/createEvent?caption=&metadata=%7B%22cameraRefs%22%3A%5B%22camera-uuid1%22%5D%7D&source=stampzilla&state=Inactive", req.URL.String())
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"server\": \"" + server.URL + "\", \"username\": \"user1\", \"password\": \"pass1\"}"))
	err := sender.Release([]string{"camera-uuid1"}, "")

	assert.NoError(t, err)
}

func TestDestinations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/ec2/getCamerasEx", req.URL.String())
		rw.Write([]byte(`[
  {
    "addParams": [
    ],
    "audioEnabled": false,
    "backupType": "CameraBackupDisabled",
    "controlEnabled": false,
    "dewarpingParams": "",
    "disableDualStreaming": false,
    "failoverPriority": "Never",
    "groupId": "",
    "groupName": "",
    "id": "{5326b322-9d30-4c61-7e53-cd83dfbb680f}",
    "licenseUsed": false,
    "logicalId": "",
    "mac": "AC-CC-8E-6E-21-DF",
    "manuallyAdded": false,
    "maxArchiveDays": -30,
    "minArchiveDays": -1,
    "model": "AXISM3037",
    "motionMask": "",
    "motionType": "MT_Default",
    "name": "AXISM3037",
    "parentId": "{c624cfea-d3e6-15a1-bb82-670fbd8a6b03}",
    "physicalId": "AC-CC-8E-6E-21-DF",
    "preferredServerId": "{00000000-0000-0000-0000-000000000000}",
    "scheduleEnabled": false,
    "scheduleTasks": [],
    "status": "Offline",
    "statusFlags": "CSF_NoFlags",
    "typeId": "{7d99e0b1-0a98-e5af-563e-51ddd66c2a4f}",
    "url": "http://198.18.0.51:554",
    "userDefinedGroupName": "",
    "vendor": "Axis"
  },
  {
    "addParams": [
    ],
    "audioEnabled": false,
    "backupType": "CameraBackupDisabled",
    "controlEnabled": false,
    "dewarpingParams": "",
    "disableDualStreaming": false,
    "failoverPriority": "Never",
    "groupId": "",
    "groupName": "",
    "id": "{669411bf-7ce1-c981-20e7-0de6d968f864}",
    "licenseUsed": false,
    "logicalId": "",
    "mac": "00-40-8C-8C-6C-A5",
    "manuallyAdded": false,
    "maxArchiveDays": -30,
    "minArchiveDays": -1,
    "model": "AXIS212PTZ",
    "motionMask": "",
    "motionType": "MT_Default",
    "name": "AXIS212PTZ",
    "parentId": "{c624cfea-d3e6-15a1-bb82-670fbd8a6b03}",
    "physicalId": "00-40-8C-8C-6C-A5",
    "preferredServerId": "{00000000-0000-0000-0000-000000000000}",
    "scheduleEnabled": false,
    "scheduleTasks": [],
    "status": "Offline",
    "statusFlags": "CSF_NoFlags",
    "typeId": "{ad3674b8-3db5-1a6f-03d4-d25ce1c4289d}",
    "url": "http://172.16.21.30:80",
    "userDefinedGroupName": "",
    "vendor": "Axis"
  },
  {
    "addParams": [
      {
        "name": "DeviceUrl",
        "value": "http://172.16.21.127:80/onvif/device_service"
      },
      {
        "name": "MaxFPS",
        "value": "25"
      },
      {
        "name": "VideoLayout",
        "value": ""
      },
      {
        "name": "bitrateInfos",
        "value": "{\"streams\":[{\"actualBitrate\":0.92454636096954346,\"actualFps\":23.571428298950195,\"averageGopSize\":4.6118630519259465e-16,\"bitrateFactor\":1,\"bitratePerGop\":\"BPG_None\",\"encoderIndex\":\"primary\",\"fps\":25,\"isConfigured\":true,\"numberOfChannels\":1,\"rawSuggestedBitrate\":4.2721843719482422,\"resolution\":\"2048x1536\",\"suggestedBitrate\":4.271484375,\"timestamp\":\"2019-10-07T12:36:46Z\"},{\"actualBitrate\":0.11209215223789215,\"actualFps\":5.7142858505249023,\"averageGopSize\":12,\"bitrateFactor\":1,\"bitratePerGop\":\"BPG_None\",\"encoderIndex\":\"secondary\",\"fps\":7,\"isConfigured\":true,\"numberOfChannels\":1,\"rawSuggestedBitrate\":0.61887556314468384,\"resolution\":\"640x480\",\"suggestedBitrate\":0.234375,\"timestamp\":\"2019-10-07T12:37:26Z\"}]}"
      },
      {
        "name": "bitratePerGOP",
        "value": "0"
      },
      {
        "name": "cameraAdvancedParams",
        "value": "{\"groups\":[{\"aux\":\"\",\"description\":\"\",\"groups\":[{\"aux\":\"\",\"description\":\"\",\"groups\":[],\"name\":\"Primary\",\"params\":[{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[],\"description\":\"\",\"group\":\"\",\"id\":\"primaryStream.codec\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Codec\",\"notes\":\"\",\"range\":\"H264\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[{\"conditions\":[{\"paramId\":\"primaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"}],\"id\":\"520f8ddb-e91a-bec7-bfb7-1cd2156855bb\",\"internalRange\":\"\",\"range\":\"176x144,320x240,640x480,800x600,1280x960,1600x1200,2048x1536\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]}],\"description\":\"\",\"group\":\"\",\"id\":\"primaryStream.resolution\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Resolution\",\"notes\":\"\",\"range\":\"\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"Are you sure you want to set defaults for this stream?\",\"dataType\":\"Button\",\"dependencies\":[],\"description\":\"\",\"group\":\"\",\"id\":\"primaryStream.resetToDefaults\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Reset to Defaults\",\"notes\":\"\",\"range\":\"\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":true,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"}]},{\"aux\":\"\",\"description\":\"\",\"groups\":[],\"name\":\"Secondary\",\"params\":[{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[],\"description\":\"\",\"group\":\"\",\"id\":\"secondaryStream.codec\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Codec\",\"notes\":\"\",\"range\":\"H264\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"}],\"id\":\"a9208df2-56c2-b333-dfe3-f80c138a7eef\",\"internalRange\":\"\",\"range\":\"176x144,320x240,640x480,800x600,1280x960,1600x1200,2048x1536\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]}],\"description\":\"\",\"group\":\"\",\"id\":\"secondaryStream.resolution\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Resolution\",\"notes\":\"\",\"range\":\"\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"1280x960\"}],\"id\":\"2e790728-1168-3dd7-861a-e446c9e103fc\",\"internalRange\":\"\",\"range\":\"192,4119\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"1600x1200\"}],\"id\":\"f6c28141-89c6-dd79-9e24-03454f640d2d\",\"internalRange\":\"\",\"range\":\"192,5629\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"176x144\"}],\"id\":\"d2e5188c-fc41-72c7-85f8-00d37fddeaed\",\"internalRange\":\"\",\"range\":\"192,272\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"2048x1536\"}],\"id\":\"79501fe1-0e96-25e2-49c1-69f9418c76df\",\"internalRange\":\"\",\"range\":\"192,7954\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"320x240\"}],\"id\":\"f5de67b4-ecc5-c51a-4ed2-bd29db83bdd1\",\"internalRange\":\"\",\"range\":\"192,591\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"640x480\"}],\"id\":\"152bab41-9329-789e-c1d7-4f86650509b9\",\"internalRange\":\"\",\"range\":\"192,1560\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"800x600\"}],\"id\":\"c64f6231-03d5-de10-1fd6-a12ae3092683\",\"internalRange\":\"\",\"range\":\"192,2133\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]}],\"description\":\"\",\"group\":\"\",\"id\":\"secondaryStream.bitrateKbps\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Bitrate\",\"notes\":\"\",\"range\":\"1,100000\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"Kbps\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"1280x960\"}],\"id\":\"13f3eb4a-3328-9dc5-4790-2fb317e29d2f\",\"internalRange\":\"\",\"range\":\"1,25\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"1600x1200\"}],\"id\":\"b371c453-cbcb-d53a-3550-8715764c6c50\",\"internalRange\":\"\",\"range\":\"1,25\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"176x144\"}],\"id\":\"8d4d760e-1e16-a045-3556-380fe220daaf\",\"internalRange\":\"\",\"range\":\"1,25\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"2048x1536\"}],\"id\":\"1f9a85de-4dfe-6aa7-7fe2-4a4a2137b663\",\"internalRange\":\"\",\"range\":\"1,25\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"320x240\"}],\"id\":\"ab6211ed-c18d-c3ee-fd67-667345f3925d\",\"internalRange\":\"\",\"range\":\"1,25\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"640x480\"}],\"id\":\"1fa321aa-2211-1c40-b9e2-4eb8431af647\",\"internalRange\":\"\",\"range\":\"1,25\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]},{\"conditions\":[{\"paramId\":\"secondaryStream.codec\",\"type\":\"value\",\"value\":\"H264\"},{\"paramId\":\"secondaryStream.resolution\",\"type\":\"value\",\"value\":\"800x600\"}],\"id\":\"a7b3b253-6ed1-aacf-2842-f5d694949873\",\"internalRange\":\"\",\"range\":\"1,25\",\"type\":\"Range\",\"valuesToAddToRange\":[],\"valuesToRemoveFromRange\":[]}],\"description\":\"\",\"group\":\"\",\"id\":\"secondaryStream.fps\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"FPS\",\"notes\":\"\",\"range\":\"1,100\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"Frames per Second\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"Are you sure you want to set defaults for this stream?\",\"dataType\":\"Button\",\"dependencies\":[],\"description\":\"\",\"group\":\"\",\"id\":\"secondaryStream.resetToDefaults\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Reset to Defaults\",\"notes\":\"\",\"range\":\"\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":true,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"}]}],\"name\":\"Video Streams Configuration\",\"params\":[]},{\"aux\":\"\",\"description\":\"\",\"groups\":[{\"aux\":\"\",\"description\":\"\",\"groups\":[],\"name\":\"Exposure\",\"params\":[{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[],\"description\":\"Exposure Mode. Enable or disable the exposure algorithm on the device.\",\"group\":\"\",\"id\":\"ieMode\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Mode\",\"notes\":\"\",\"range\":\"Auto,Manual\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[],\"description\":\"The exposure priority mode.\",\"group\":\"\",\"id\":\"iePriority\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Priority\",\"notes\":\"\",\"range\":\"LowNoise,FrameRate\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"}]},{\"aux\":\"\",\"description\":\"\",\"groups\":[],\"name\":\"Focus\",\"params\":[{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[],\"description\":\"Mode of Auto Focus\",\"group\":\"\",\"id\":\"ifAuto\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Auto Focus\",\"notes\":\"\",\"range\":\"Auto,Manual\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"}]},{\"aux\":\"\",\"description\":\"\",\"groups\":[],\"name\":\"White Balance\",\"params\":[{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[],\"description\":\"White Balance.\",\"group\":\"\",\"id\":\"iwbMode\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Mode\",\"notes\":\"\",\"range\":\"Auto,Manual\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[],\"description\":\"\",\"group\":\"\",\"id\":\"iwbYrGain\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Yr Gain\",\"notes\":\"\",\"range\":\"0,100\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[],\"description\":\"\",\"group\":\"\",\"id\":\"iwbYbGain\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Yb Gain\",\"notes\":\"\",\"range\":\"0,100\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"}]}],\"name\":\"Imaging\",\"params\":[{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[],\"description\":\"Image brightness.\",\"group\":\"\",\"id\":\"iBri\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Brightness\",\"notes\":\"\",\"range\":\"0,100\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[],\"description\":\"Color saturation of the image.\",\"group\":\"\",\"id\":\"iCS\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Color Saturation\",\"notes\":\"\",\"range\":\"0,100\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[],\"description\":\"Contrast of the image.\",\"group\":\"\",\"id\":\"iCon\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Contrast\",\"notes\":\"\",\"range\":\"0,100\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Enumeration\",\"dependencies\":[],\"description\":\"Infrared Cutoff Filter settings.\",\"group\":\"\",\"id\":\"iIrCut\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Ir Cut Filter Mode\",\"notes\":\"\",\"range\":\"On, Off, Auto\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"\",\"dataType\":\"Number\",\"dependencies\":[],\"description\":\"Sharpness of the Video image.\",\"group\":\"\",\"id\":\"iSha\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Sharpness\",\"notes\":\"\",\"range\":\"0,100\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"}]},{\"aux\":\"\",\"description\":\"\",\"groups\":[],\"name\":\"Maintenance\",\"params\":[{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"Are you sure you want to reboot the device?\",\"dataType\":\"Button\",\"dependencies\":[],\"description\":\"This operation reboots the device.\",\"group\":\"\",\"id\":\"mReboot\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"System Reboot\",\"notes\":\"\",\"range\":\"\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"Are you sure you want to reset all settings (except network) to default?\",\"dataType\":\"Button\",\"dependencies\":[],\"description\":\"This operation reloads all parameters on the device to their factory default values, except basic network settings like IP address, subnet and gateway or DHCP settings.\",\"group\":\"\",\"id\":\"mSoftReset\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Soft Factory Reset\",\"notes\":\"\",\"range\":\"\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"},{\"aux\":\"\",\"bindDefaultToMinimum\":false,\"confirmation\":\"Are you sure you want to reset all settings (including network) to default?\",\"dataType\":\"Button\",\"dependencies\":[],\"description\":\"This operation reloads all parameters on the device to their factory default values.\",\"group\":\"\",\"id\":\"mHardReset\",\"internalRange\":\"\",\"keepInitialValue\":false,\"name\":\"Hard Factory Reset\",\"notes\":\"\",\"range\":\"\",\"readCmd\":\"\",\"readOnly\":false,\"resync\":false,\"showRange\":false,\"tag\":\"\",\"unit\":\"\",\"writeCmd\":\"\"}]}],\"name\":\"primaryStreamConfiguration, secondaryStreamConfiguration, ONVIF\",\"packet_mode\":true,\"unique_id\":\"971009D8-384D-43C1-85BF-44A2B99253D5, 971009D8-384D-43C1-85BF-44A2B99253D5, 53148996-d4d8-4e90-985e-24c6c63c5150\",\"version\":\"1.0, 1.0, 1\"}"
      },
      {
        "name": "dontRecordPrimaryStream",
        "value": "0"
      },
      {
        "name": "dontRecordSecondaryStream",
        "value": "0"
      },
      {
        "name": "firmware",
        "value": "0100"
      },
      {
        "name": "hasDualStreaming",
        "value": "1"
      },
      {
        "name": "ioOverlayStyle",
        "value": "Form"
      },
      {
        "name": "ioSettings",
        "value": "[]"
      },
      {
        "name": "isAudioSupported",
        "value": "1"
      },
      {
        "name": "mediaCapabilities",
        "value": "{\"hasAudio\":false,\"hasDualStreaming\":false,\"streamCapabilities\":[{\"key\":\"primary\",\"value\":{\"defaultBitrateKbps\":0,\"defaultFps\":0,\"maxBitrateKbps\":7954,\"maxFps\":25,\"minBitrateKbps\":192}},{\"key\":\"secondary\",\"value\":{\"defaultBitrateKbps\":0,\"defaultFps\":0,\"maxBitrateKbps\":1560,\"maxFps\":25,\"minBitrateKbps\":192}}]}"
      },
      {
        "name": "mediaStreams",
        "value": "{\"streams\":[{\"codec\":28,\"customStreamParams\":{\"profile-level-id\":\"420032\",\"sprop-parameter-sets\":\"J0IAMpY1AEABg03BQYFQAAA+kAAOpgC+oA==,KM4EYg==\"},\"encoderIndex\":0,\"resolution\":\"2048x1536\",\"transcodingRequired\":false,\"transports\":[\"rtsp\",\"hls\"]},{\"codec\":28,\"customStreamParams\":{\"profile-level-id\":\"4d001f\",\"sprop-parameter-sets\":\"J00AH5Y1AUB7TcFBgVAAAD6QAA6mAb6g,KO4EYg==\"},\"encoderIndex\":1,\"resolution\":\"640x480\",\"transcodingRequired\":false,\"transports\":[\"rtsp\",\"hls\"]},{\"codec\":0,\"customStreamParams\":{},\"encoderIndex\":-1,\"resolution\":\"*\",\"transcodingRequired\":true,\"transports\":[\"rtsp\",\"mjpeg\",\"webm\"]}]}"
      },
      {
        "name": "motionStream",
        "value": ""
      },
      {
        "name": "overrideAr",
        "value": ""
      },
      {
        "name": "ptzCapabilities",
        "value": "987767"
      },
      {
        "name": "ptzPresets",
        "value": "{}"
      },
      {
        "name": "rotation",
        "value": ""
      },
      {
        "name": "rtpTransport",
        "value": ""
      },
      {
        "name": "streamUrls",
        "value": "{\n    \"1\": \"rtsp://172.16.21.127/live.sdp\",\n    \"2\": \"rtsp://172.16.21.127/live2.sdp\"\n}\n"
      }
    ],
    "audioEnabled": false,
    "backupType": "CameraBackupDefault",
    "controlEnabled": true,
    "dewarpingParams": "{\"enabled\":true,\"fovRot\":0,\"hStretch\":1,\"radius\":0.62,\"viewMode\":\"0\",\"xCenter\":0.48999999999999999,\"yCenter\":0.49934895833333326}",
    "disableDualStreaming": false,
    "failoverPriority": "Medium",
    "groupId": "",
    "groupName": "",
    "id": "{cfad7e8a-fe12-9f4b-3b0d-359aec82725c}",
    "licenseUsed": true,
    "logicalId": "",
    "mac": "00-02-D1-4D-4E-D9",
    "manuallyAdded": true,
    "maxArchiveDays": -30,
    "minArchiveDays": -1,
    "model": "CC8370-HV",
    "motionMask": "0,0,0,44,19;0,8,19,8,6;0,8,25,36,1;0,0,26,44,6;5,16,19,28,6;5,0,19,8,7",
    "motionType": "MT_SoftwareGrid",
    "name": "CC8370-HV",
    "parentId": "{c624cfea-d3e6-15a1-bb82-670fbd8a6b03}",
    "physicalId": "00-02-D1-4D-4E-D9",
    "preferredServerId": "{c624cfea-d3e6-15a1-bb82-670fbd8a6b03}",
    "scheduleEnabled": false,
    "scheduleTasks": [
      {
        "afterThreshold": 5,
        "beforeThreshold": 5,
        "bitrateKbps": 0,
        "dayOfWeek": 1,
        "endTime": 86400,
        "fps": 10,
        "recordAudio": false,
        "recordingType": "RT_Never",
        "startTime": 0,
        "streamQuality": "highest"
      },
      {
        "afterThreshold": 5,
        "beforeThreshold": 5,
        "bitrateKbps": 0,
        "dayOfWeek": 2,
        "endTime": 86400,
        "fps": 10,
        "recordAudio": false,
        "recordingType": "RT_Never",
        "startTime": 0,
        "streamQuality": "highest"
      },
      {
        "afterThreshold": 5,
        "beforeThreshold": 5,
        "bitrateKbps": 0,
        "dayOfWeek": 3,
        "endTime": 86400,
        "fps": 10,
        "recordAudio": false,
        "recordingType": "RT_Never",
        "startTime": 0,
        "streamQuality": "highest"
      },
      {
        "afterThreshold": 5,
        "beforeThreshold": 5,
        "bitrateKbps": 0,
        "dayOfWeek": 4,
        "endTime": 86400,
        "fps": 10,
        "recordAudio": false,
        "recordingType": "RT_Never",
        "startTime": 0,
        "streamQuality": "highest"
      },
      {
        "afterThreshold": 5,
        "beforeThreshold": 5,
        "bitrateKbps": 0,
        "dayOfWeek": 5,
        "endTime": 86400,
        "fps": 10,
        "recordAudio": false,
        "recordingType": "RT_Never",
        "startTime": 0,
        "streamQuality": "highest"
      },
      {
        "afterThreshold": 5,
        "beforeThreshold": 5,
        "bitrateKbps": 0,
        "dayOfWeek": 6,
        "endTime": 86400,
        "fps": 10,
        "recordAudio": false,
        "recordingType": "RT_Never",
        "startTime": 0,
        "streamQuality": "highest"
      },
      {
        "afterThreshold": 5,
        "beforeThreshold": 5,
        "bitrateKbps": 0,
        "dayOfWeek": 7,
        "endTime": 86400,
        "fps": 10,
        "recordAudio": false,
        "recordingType": "RT_Never",
        "startTime": 0,
        "streamQuality": "highest"
      }
    ],
    "status": "Offline",
    "statusFlags": "CSF_NoFlags",
    "typeId": "{9a55ee6b-a595-5807-a5ba-d4aff697dc12}",
    "url": "http://172.16.21.127:80/onvif/device_service",
    "userDefinedGroupName": "",
    "vendor": "VIVOTEK"
  }
]`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"server\": \"" + server.URL + "\", \"username\": \"user1\", \"password\": \"pass1\"}"))
	d, err := sender.Destinations()
	assert.Equal(t, d, map[string]string{"{669411bf-7ce1-c981-20e7-0de6d968f864}": "AXIS212PTZ", "{cfad7e8a-fe12-9f4b-3b0d-359aec82725c}": "CC8370-HV", "{5326b322-9d30-4c61-7e53-cd83dfbb680f}": "AXISM3037"})
	assert.NoError(t, err)
}
