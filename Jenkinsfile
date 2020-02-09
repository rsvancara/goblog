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
                sh 'cp /home/artifacts/geoip/*.mmdb db/'
                sh 'docker build --no-cache -t rsvancara/goblog:jenkins .'
            }
        }
        stage('Build Release') {
            //when { buildingTag() }
            steps {
                latestTag = sh(returnStdout:  true, script: "git tag --sort=-creatordate | head -n 1").trim()
                sh 'cp /home/artifacts/geoip/*.mmdb db/'
                sh 'docker build --no-cache -t rsvancara/goblog:${latestTag} .'
                sh 'docker push rsvancara/goblog:${latestTag}'
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
