# GoM2xLogger
Log air quality data to the M2X service, written in Go

Just need to go build this and then edit cron to auto run this every X minutes, here I'm using 4 minutes to stay under the 100k values/month free m2x account limit
XM2XKEY=XXXXXXXXXXXXXXXXXXX
*/4 * * * * /home/pi/work/src/github.com/brandonagr/gom2xlogger/gom2xlogger >> /var/log/gom2xlogger.log 2>&1
30 1 * * 7 /sbin/shutdown -r now
