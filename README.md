# BarkBridge

Bridge your GSM modems to any HTTP endpoint.

## Description

You run a service that  
• Detects USB GSM modems hot-plugged under `/dev/ttyUSB*`  
• Listens for SMS and missed calls  
• Posts each event as JSON to your endpoint  

## Prerequisites

• Go 1.20+ installed  
• USB GSM modems supported by Linux serial drivers  
• An HTTP server ready to receive POSTs  

## Installation

1. Clone the repo  
   ```bash
   git clone https://github.com/huskycodesoy/BarkBridge.git
   cd BarkBridge
   ```
2. Fetch dependencies  
   ```bash
   go get github.com/tarm/serial@latest
   go mod tidy
   ```
3. Build the binary  
   ```bash
   go build -o barkbridge
   ```

## Configuration

Set at runtime via flags:

• `-endpoint`  
  URL to receive JSON payloads  

• `-baud`  
  Serial baud rate (default: 115200)  

## Usage

Run BarkBridge:

```bash
./barkbridge \
  -endpoint https://your.endpoint/receive \
  -baud 115200
```

It scans for modems every minute.  
It logs attach, detach, and errors to stdout.

## JSON Payload

Each event sends:

```json
{
  "type": "sms" | "call",
  "modem": "/dev/ttyUSB0",
  "timestamp": "2025-07-15T12:00:00Z",
  "from": "+1234567890",
  "text": "Hello"      // only for SMS
}
```

## License

MIT License

Copyright (c) 2025 HuskyCodes Oy

Permission is hereby granted, free of charge, to any person obtaining a copy  
of this software and associated documentation files (the "Software"), to deal  
in the Software without restriction, including without limitation the rights  
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell  
copies of the Software, and to permit persons to whom the Software is  
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in  
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR  
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,  
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE  
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER  
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,  
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN  
THE SOFTWARE.