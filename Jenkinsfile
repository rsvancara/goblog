#!/usr/bin/env groovy

pipeline {
    agent any

    options {
        ansiColor(colorMapName: 'XTerm')
        timestamps()
    }
    stages {
        stage('Build Release') {
            def tag = sh(returnStdout: true, script: "git tag --contains | head -1").trim()
            if (tag) {
            steps {
                sh 'cp /home/artifacts/geoip/*.mmdb db/'
                sh 'docker build --no-cache -t rsvancara/goblog:${tag} .'
                sh 'docker push rsvancara/goblog:${tag}'
            }
            }
        }
        stage('Publish visualintrigue') {

            steps {
                //withDockerRegistry([ credentialsId: "dockerhub", url: "https://docker.io/" ]) {
                //    sh 'kubectl get pods -o wide -n dev'
                //}
                sh 'kubectl get pods -o wide -n dev'
            }
        }
    }
}
