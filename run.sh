#!/bin/bash

Xephyr :5 -terminate -screen 1910x1030 &
sleep 1
DISPLAY=:5 ./loomy
