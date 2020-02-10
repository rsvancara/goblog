#!/usr/bin/env groovy

pipeline {
    agent any

    options {
        ansiColor(colorMapName: 'XTerm')
        timestamps()
    }
    stages {
        stage('Build Release') {

            steps {
                sh 'scripts/build.sh'
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
