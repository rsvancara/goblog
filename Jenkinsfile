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
                
            }
        }
        stage('Build') {
            steps {
                sh 'docker build --no-cache -t rsvancara/goblog:${TAG_NAME} .'
            }
        }

        stage('Publish') {
            when { buildingTag() }
            steps {
                withDockerRegistry([ credentialsId: "jenkins-inf", url: "https://docker.util.pages/" ]) {
                    sh 'docker push docker.util.pages/inf/dabloog:${TAG_NAME}'
                }
            }
        }
    }
}
