# Streaming service

## Requirements
Non negotiables for service:
 - Can watch videos
 - Can upload videos

Add ons:
 - Video recommendations (need tags for this and user data like, likes and following)
 - Views and Like tracking
 - Searching

## Introduction
Users can upload a video. Uploaded video is processed by FFMPEG to go to 480p (later we'll add the ability to save multiple qualities). File is then processed to a COS bucket.
Do we save multiple quality formats in the bucket or do we save one of the formats and transcribe it to another quality level during the streaming?
- Multiple files for different quality levels -> higher storage costs, less time spend on processing on server.
- One file transcribing on server -> More load on server and higher latency, save costs. This does not scale but does work for a start.

## Routes
GET /video-meta-data/{id}
  - Get video info:
    - Title
    - Description

GET /video/{id}?quality=480p&offset=100
  - Get the video bytes for streaming. AKA load file from COS and stream the file bytes over to client.
  - Can submit quality. When not submitted use 480p. For start we always set videos to 480p
  - Can submit offset. Can choose where to start the streaming of the video from. Default is 0. Measured in bytes / seconds (find out which easier)

POST /video
 - Create new video
 - body
  - file
  - title
  - description (not required)

## Tables / entities
Video table
  - id PRIMARY KEY
  - title VARCHAR(200) NOT NULL
  - description MEDIUMTEXT
  - cos_bucket_file_name VARCHAR(200)

