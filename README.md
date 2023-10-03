# ip2tg

This piece of code is complemented by 'ip2mqtt' which checks your IP address from an online portal
i.e. https://api.ipify.org?format=json
Format:
{"ip":"new_ip_address"}
and pushes it to a MQTT bus.

On first run, ip2tg sets a default IP address to 127.0.0.1 and subscribes to a MQTT topic where it
waits for an update.
Once it has an updated IP, it sends it to a Telegram recepient (user or channel) and as well as 
updates it in your Dynu DNS so you don't ever have to do it manually. Your Telegraam will get a status update for this
too in form of "200 OK" (if successful) or whatever HTTP error is returned

So the obvious pre-reqs are:
- ip2mqtt (or a simple linux curl script getting your IP and publishing it to MQTT)
- MQTT (I used Mosquitto but feel free to explore other messaging servers which can perform a similar function)
- Telegram user or channel ID to where message will be sent
- Telegram Bot and your Bot API key
- Dynu account and API key

All config items come in as environment variables so you can deploy to physical, virtual, K8s windows, linux, whatever...
Some of them need to be changed, others are up to you!
Here they are:

MYIP2TG_APIKEY="<Your Telegram API Key>"
MYIP2TG_TARGET="<Your Telegram user or channel ID"
MYIP2TG_IPSITE="https://api.ipify.org?format=json"
MYIP2TG_MQTT="127.0.0.1" 
MYIP2TG_MQTTPORT="1883"
MYIP2TG_MQTT_USER="<MQTT user>"
MYIP2TG_MQTT_PWD="<MQTT user password"
MYIP2TG_MQTT_TOPIC="topic/homeip"
MYIP2TG_MQTT_CLID="ip2tg_client"
MYIP2TG_DDNSAPI="https://api.dynu.com/v2/dns/<your host ID>" // you can get this ID by doing a curl GET on the above API
MYIP2TG_DDNSKEY="Your DDNS API Key"
MYIP2TG_DOMNAME="FQDN of your host in DDNS"
MYIP2TG_DOMGRP="Group that you created for your host (tool may break without setting this)"
