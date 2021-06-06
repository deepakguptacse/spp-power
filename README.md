# spp-power
Checks the electricity power state of the society whether the source is main or generator.

## How does it work?
It was based on two observations:
- When the source changes from MAIN to GENERATOR, the power of the flat goes away for more than 15 seconds.
- When the source changes from GENERATOR to MAIN, the power of the flat goes away for less than 15 seconds.

This is run on a laptop which is always hooked on to the charging cable. This program monitors the battery charging/discharging events and infers the current source of power based on the two abovementioned observations.

The program exposes a HTTPS endpoint on port 443, which is:
- Green if the source of power is MAIN.
- Red if the source of power is GENERATOR.

## How to run the program locally:

### Run the server locally:
```
sudo go run main/main.go <current-state>
```
- `<current-state>` is main if the source of power is MAIN
- `<current-state>` is gen if the source of power is GENERATOR

Note: `sudo` is required because of the predefined port 80.

Check the current electricity source:
```
https://localhost:443
```

### For client side changes:
```
firebase deploy
```
Please ask for access to the firebase project: deepakguptacse@gmail.com

