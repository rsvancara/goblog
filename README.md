# GoBlog - Go Based Blogging Software


# Instructions

## Installation, Instructions

### Docker

### Command Line Commando!

## Using the file API


### File Upload using curl
Files must be in the XML format.  If not the file may upload, but results will not be processed or put into the database.  

```bash
curl -F "data=@nmap.results.xml" http://localhost:5000/api/v1/putnmap/your_session_goes_here
```

Note: Please replace the string, your_session_goes_here with a unique session name.  Also session names must be unique from one file upload to the next.

### Using the Web Interface


# NMAP File Format Choice


# Assumptions


# Additional Thoughts


# Database Schema

