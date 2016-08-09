# GoM2xLogger
Log air quality data to the M2X service, written in Go. 

Just need to go build this and then edit cron to auto run this every X minutes, here I'm using 4 minutes to stay under the 100k values/month free m2x account limit
  XM2XDEVICE=XXXXXXXXXXXXXXXXXXX
  XM2XKEY=XXXXXXXXXXXXXXXXXXX
  */4 * * * * /home/pi/work/src/github.com/brandonagr/gom2xlogger/gom2xlogger >> /var/log/gom2xlogger.log 2>&1
  30 1 * * 7 /sbin/shutdown -r now


## Initial pi setup
To go from a fresh install of Raspian Jessie Lite

  sudo passwd pi
	change to new password

  sudo nano /etc/default/keyboard
	change gb to us

  sudo iwlist wlan0 scan | more
	find wifi to setup

  sudo nano /etc/wpa_supplicant/wpa_supplicant.conf 
  network={
  	ssid="..."
  	psk="..."
  }

  sudo reboot

  sudo apt-get update
  sudo apt-get ugprade
  sudo apt-get install git

  wget https://storage.googleapis.com/golang/go1.6.3.linux-armv6l.tar.gz
  sudo tar -C /usr/local -xzf go1.6.3.linux-armv6l.tar.gz
  nano $HOME/.profile
  	export GOROOT=/usr/local/go
  	export PATH=$PATH:/usr/local/go/bin
  	export GOPATH=$HOME/work
  cd ~ && mkdir work

  sudo apt-get install libasound2-dev
  go get github.com/cocoonlife/goalsa

  arecord -l
  	should output the usb mic, card 1 device 0 which is hw:1,0

## Compile Adafruit DHT library
This is a python / c library, and we just need the c part to interface with using cgo instead of python

  git clone https://github.com/adafruit/Adafruit_Python_DHT.git
  cd Adafruit_Python_DHT/source
  gcc Raspberry_Pi_2/pi_2_dht_read.c Raspberry_Pi_2/pi_2_mmio.c common_dht_read.c -std=gnu99 -shared -o libdht.so
  sudo cp libdht.so /usr/lib/libdht.so
