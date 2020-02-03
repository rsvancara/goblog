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
            steps {
                sh 'docker build --no-cache -t rsvancara/goblog:release .'
                sh 'docker push rsvancara/goblog:release'
            }
        }
        stage('Publish visualintrigue') {
            //when { buildingTag() }
            steps {
                //withDockerRegistry([ credentialsId: "dockerhub", url: "https://docker.io/" ]) {
                //    sh 'kubectl get pods -o wide -n dev'
                //}
                sh 'kubectl get pods -o wide -n dev'
            }
        }
    }
}
