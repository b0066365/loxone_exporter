#!/usr/bin/env python
# -*- coding: utf-8 -*-

import json
import requests
import time
import random
import urllib3
import sys
import base64
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

#-------------------------------------------------
def GetStromverbrauch(headers):
    loxone_devices = {'SSA-01 Leistung', 'SSA-02 Leistung', 'SSA-03 Leistung', 'SSA-04 Leistung'}
    metrics = {}

    for loxone_device in loxone_devices:
        get_response = requests.get("http://10.0.0.77/dev/sps/io/"+loxone_device+"/state", headers=headers, verify=False)

        if get_response.status_code == 200:
            value=get_response.text.split('" ')
            value = value[2].lstrip('value="')
            value = value[:-2]
            #value = value*1000
            loxone_device = loxone_device.split(" ")
            metrics[loxone_device[0]] = value

        time.sleep(0.5)
    return metrics

def GetStellmotor(headers):
    loxone_devices = {'SAA-01', 'SAA-02', 'SAA-03', 'SAA-04'}
    metrics = {}

    for loxone_device in loxone_devices:
        get_response = requests.get("http://10.0.0.77/dev/sps/io/"+loxone_device+"/state", headers=headers, verify=False)

        if get_response.status_code == 200:
            value=get_response.text.split('" ')
            value = value[2].lstrip('value="')
            metrics[loxone_device] = value

        time.sleep(0.5)
    return metrics

def GetHumidity(headers):
    loxone_devices = {'TFS-01 Luftfeuchte'}
    metrics = {}

    for loxone_device in loxone_devices:
        get_response = requests.get("http://10.0.0.77/dev/sps/io/"+loxone_device+"/state", headers=headers, verify=False)

        if get_response.status_code == 200:
            value=get_response.text.split('" ')
            value = value[2].lstrip('value="')
            value = value[:-1]
            loxone_device = loxone_device.split(" ")
            metrics[loxone_device[0]] = value

        time.sleep(0.5)
    return metrics

def GetTemperatures(headers):
    loxone_devices = {'SSA-01 Temperatur', 'SSA-02 Temperatur','SSA-03 Temperatur','SSA-04 Temperatur','TFS-01 Temperatur'}
    metrics = {}

    for loxone_device in loxone_devices:
        get_response = requests.get("http://10.0.0.77/dev/sps/io/"+loxone_device+"/state", headers=headers, verify=False)

        if get_response.status_code == 200:
            value=get_response.text.split('" ')
            value = value[2].lstrip('value="')
            value = value[:-2]
            loxone_device = loxone_device.split(" ")
            metrics[loxone_device[0]] = value

        time.sleep(0.5)
    return metrics

def PostDataToInfluxDB(metrics, datatype):
    for metric in metrics:
        post_response = requests.post("http://10.0.0.51:8086/write?db=loxone", data=datatype+",source=miniserver,device="+metric+" value="+str(metrics[metric])+"")
        print metric+": "+str(metrics[metric])+" - "+str(post_response)
        time.sleep(0.5)

    metrics = {}

#--------- MAIN -------------

if __name__ == '__main__':
    headers = { 'Authorization' : 'Basic ZGlyazokY2lzY28xMg==' }

    PostDataToInfluxDB(GetTemperatures(headers), "Temp")
    PostDataToInfluxDB(GetHumidity(headers), "Humidity")
    PostDataToInfluxDB(GetStellmotor(headers), "Stellmotor")
    PostDataToInfluxDB(GetStromverbrauch(headers), "Stromverbrauch")
