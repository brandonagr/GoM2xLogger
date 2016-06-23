#include "DHT.h" // from https://github.com/adafruit/DHT-sensor-library

DHT dht(13, DHT22);

// the setup function runs once when you press reset or power the board
void setup() {
  Serial.begin(9600);

  dht.begin();
  
  // initialize digital pin 13 as an output.
  pinMode(13, OUTPUT);
}

// the loop function runs over and over again forever
void loop() {

  digitalWrite(13, HIGH);   // turn the LED on (HIGH is the voltage level)

  // wait 1 second
  delay(1000);

  float temperature = dht.readTemperature(true, false);
  float humidity = dht.readHumidity();

  if (isnan(temperature) || isnan(humidity)) {
    return;
  }

  Serial.print(temperature);
  Serial.print(" ");
  Serial.print(humidity);
  Serial.println();

  digitalWrite(13, LOW);    // turn the LED off by making the voltage LOW
}
