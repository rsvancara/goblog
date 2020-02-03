#!/usr/bin/env groovy

pipeline {
    agent any

    options {
        ansiColor(colorMapName: 'XTerm')
        timestamps()
    }
    stages {
        stage('Test') {
            steps {
                sh 'docker build --no-cache -t rsvancara/goblog:jenkins .'
            }
        }
        stage('Build Release') {
            when { buildingTag() }
            steps {
                sh 'docker build --no-cache -t rsvancara/goblog:${TAG_NAME} .'
            }
        }
        stage('Publish visualintrigue') {
            when { buildingTag() }
            steps {
                withDockerRegistry([ credentialsId: "dockerhub", url: "https://docker.io/" ]) {
                    sh 'docker push rsvancara/goblog:vi-${TAG_NAME}'
                }
            }
        }
        stage('Publish dyitinytrailer') {
            when { buildingTag() }
            steps {
                withDockerRegistry([ credentialsId: "dockerhub", url: "https://docker.io/" ]) {
                    sh 'docker push rsvancara/goblog:dyi-${TAG_NAME}'
                }
            }
        }
    }
}
