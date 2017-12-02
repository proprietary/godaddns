# Godaddns

A small program that implements "dynamic DNS" or "ddns" for domains hosted on Godaddy.

## Requirements

You will need to [get an API key and API secret](https://developer.godaddy.com/keys/) from Godaddy. Make sure it's a production key, not a testing key.

## Download

## Usage

`$ ./godaddns -key="my api key" -secret="my api secret" -domain="example.com" -polling="(optional) polling interval in seconds; defaults to 360 seconds" -subdomain="(optional) if your target domain is subdomain.example.com, put 'subdomain' here; defaults to '@'" -log "(optional) path to log file; defaults to stdout"`