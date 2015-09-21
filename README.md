# ComposeHub CLI

A tool to publish, search and install apps using docker-compose. You can also search on ComposeHub using the web UI online https://composehub.com

The Docker hub is awesome to search for great images to build your own apps.
However, for existing apps such as wordpress or gitlab, you need more than one container and linking them all together is no fun as it requires typing never-ending docker CLI commands. So you end up googling for a docker-compose.yml to solve your problem and end up copy/pasting the result in your own docker-compose.yml file, crossing fingers it works out. 
Composehub solves this problem by providing an easy way to search for docker-compose apps stored on git repos. You can also use it to publish your own public or private apps. Give it a try!


## Overview

* Publish apps
* Search for apps
* Install apps in seconds
* Run apps on the fly

# Getting Started

## Install on OSX:

```
curl -L https://composehub.com/install/darwin > /usr/local/bin/ch
```

## Install on Linux:

```
curl -L https://composehub.com/install/linux > /usr/local/bin/ch
```

## Install on Windows:

```
curl -L https://composehub.com/install/windows > /usr/local/bin/ch
```


## Install from source (requires a go install):

```
go get -u github.com/composehub/cli
```

## Documentation 


## Table of Contents

- [Search apps](#search-apps)
- [Install apps](#install-apps)
- [Run apps](#run-apps)
- [Manage your user account (optional)](#manage-account)
  - [Create an account](#create-an-account)
  - [Update an account](#update-an-account)
  - [Reset password](#reset-password)
- [Publish apps](#publish-apps)
  - [Command](#command)
  - [Config format](#config-format)


## Search apps

```
ch search gitlab
```

This will return a list of packages having gitlab in their name or description, ordered by most downloaded.

## Install apps

```
mkdir gitlab && cd gitlab
ch install gitlab
```

Before installing an app, make sure your current directory is empty. Installing the app will clone the repo containing the docker-compose.yml file. Once the installation is done, just run the usual ```docker-compose up```. If there are additional commands to execute before, they will be shown at the end of the install.

## Run apps

```
mkdir wordpress && cd wordpress
ch run wordpress
```

This will install wordpress in the current directory and run it automatically.

## Manage your user account (optional)

You only need an account if you want to publish your own apps.

### Create an account

```
ch adduser
```

You'll be asked to enter email, handle and password.

### Update an account

```
ch updateuser
```

Use this to update any of your user information.


### Reset password

```
ch resetpassword
```

Use this if you've forgotten your password.

## Publish apps

### Command

```
ch init
```

This will create a composehub.yml file in the current directory.

### Config format

```yml
---
name: package-name
blurb: 80 chars line blurb
description: |
  longer description
email: foo@bar.com
repo_url: http://github.com/foo/bar
tags: tag1,tag2
private: false
cmd: docker-compose up
```

The description will be displayed at the end of the install process of your package, use it to document any post-install required tasks. ```cmd``` is the command that will be ran when the user executes ```ch run <package>```, it is optional and can just be left blank. ```private``` if you set private to true, only you will be able to install the app.
