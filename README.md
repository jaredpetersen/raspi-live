# raspi-live
Raspi-live is a Node.js Express webserver that takes streaming video from the Raspberry Pi Camera module and makes it available on the web via [HLS](https://en.wikipedia.org/wiki/HTTP_Live_Streaming).

## Usage
Run the usual `npm start` after downloading all of the dependencies. The application will start streaming encrypted video from the Raspberry Pi camera module via the `/camera/livestream.m3u8` resource on port 8080.

## Installation
Raspi-live only supports video streaming from the Raspberry Pi camera module. Here's a the official documentation on how to connect and configure it: https://www.raspberrypi.org/documentation/usage/camera/.

Since this is a Node.js application that has Node.js dependencies, run `npm install` as well.

Raspi-live uses FFmpeg, a video conversion command-line utility, to process the streaming H.264 video that the Raspberry Pi camera module outputs. Here's how to install it on your Raspberry Pi:

1. Download and configure FFmpeg via:
```
git clone https://github.com/FFmpeg/FFmpeg.git
cd FFmpeg
sudo ./configure --arch=armel --target-os=linux --enable-gpl --enable-omx --enable-omx-rpi --enable-nonfree
```
2. If you're working with a Raspbery Pi 2 or 3, then run `sudo make -j4` to build FFmpeg. If you're working with a Raspberry Pi Zero, then run `sudo make`.
3. Install FFmpeg via `sudo make install` regardless of the model of your Raspberry Pi.
4. Delete the FFmpeg directory that was created during the git clone process in Step 1. FFmpeg has already been installed and the directory is no longer needed.

## Video Stream Playback
Raspi-live is only concerned with streaming video from the camera module and is not opinionated about how the stream should be played. Any HLS-capable player should be compatible.

However, the [raspi-live-dash](https://github.com/jaredpetersen/raspi-live-dash) project is available for those looking for an opinion.

## Why HLS instead of MPEG DASH?
While MPEG DASH is a more open standard for streaming video over the internet, HLS is more widely adopted and supported. FFmpeg, chosen for its near-ubiquitousness in the video processing space, technically supports both formats but to varying degrees. The MPEG DASH format is not listed in the [FFmpeg format documentation](https://www.ffmpeg.org/ffmpeg-formats.html) and there are minor performance issues in terms of framerate drops and and degraded image quality with FFmpeg's implementation. While these performance issues can likely be configured away with more advanced options, HLS is simply easier to implement out of the box.
