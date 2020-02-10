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
        stage('Publish visualintrigue') {
            steps {
                sh 'scripts/release.sh'
            }
        }
    }
}
