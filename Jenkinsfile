pipeline {
    agent any
    
    environment {
        // 阿里云容器镜像服务配置。
        ACR_SERVER = 'registry.cn-hangzhou.aliyuncs.com'
        ACR_NAMESPACE = 'nginx-vmware'
        ACR_REPOSITORY = 'jenkins-test'
        ACR_USERNAME = 'xiajiajia'
        ACR_PASSWORD = '990111xjj.'
        // GitHub配置
        GITHUB_REPO = 'https://github.com/xiejiajia01/test-jenkins'
        // Docker服务配置
        DOCKER_HOST = 'tcp://docker-service:2375'  // 使用K8s服务名称
    }
    
    stages {
        stage('Checkout') {
            steps {
                cleanWs()
                git credentialsId: '10267abf-ea8a-46e6-bb6a-7ef9b289727f',
                    branch: 'main',
                    url: "${GITHUB_REPO}"
            }
        }
        
        stage('Build and Push Docker Image') {
            steps {
                sh '''
                    # 设置Docker守护进程地址
                    export DOCKER_HOST=${DOCKER_HOST}
                    
                    # 登录阿里云容器镜像服务
                    docker login ${ACR_SERVER} -u ${ACR_USERNAME} -p ${ACR_PASSWORD}
                    
                    # 构建Docker镜像
                    docker build -t ${ACR_SERVER}/${ACR_NAMESPACE}/${ACR_REPOSITORY}:${BUILD_NUMBER} .
                    
                    # 添加latest标签
                    docker tag ${ACR_SERVER}/${ACR_NAMESPACE}/${ACR_REPOSITORY}:${BUILD_NUMBER} ${ACR_SERVER}/${ACR_NAMESPACE}/${ACR_REPOSITORY}:latest
                    
                    # 推送镜像
                    docker push ${ACR_SERVER}/${ACR_NAMESPACE}/${ACR_REPOSITORY}:${BUILD_NUMBER}
                    docker push ${ACR_SERVER}/${ACR_NAMESPACE}/${ACR_REPOSITORY}:latest
                    
                    # 清理本地镜像
                    docker rmi ${ACR_SERVER}/${ACR_NAMESPACE}/${ACR_REPOSITORY}:${BUILD_NUMBER}
                    docker rmi ${ACR_SERVER}/${ACR_NAMESPACE}/${ACR_REPOSITORY}:latest
                '''
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
    }
}
