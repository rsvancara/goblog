#!/usr/bin/env groovy

pipeline {
    agent any

    options {
        ansiColor(colorMapName: 'XTerm')
        timestamps()
    }

    stages {
        stage('Build') {
            steps {
                sh 'docker build .'
            }
        }
        stage('Publish') {
            when { buildingTag() }
            steps {
                sh "./build-and-push.sh docker.util.pages/inf/dabloog ${TAG_NAME}"
            }
        }
    }
}