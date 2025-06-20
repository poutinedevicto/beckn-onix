import * as cdk from 'aws-cdk-lib';
import * as eks from 'aws-cdk-lib/aws-eks';
import * as helm from 'aws-cdk-lib/aws-eks';
import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { ConfigProps } from './config';
import * as crypto from 'crypto';


interface HelmCommonServicesStackProps extends StackProps {
    config: ConfigProps;
    eksCluster: eks.Cluster;
    service: string,
}

export class HelmCommonServicesStack extends Stack {
    constructor(scope: Construct, id: string, props: HelmCommonServicesStackProps) {
        super(scope, id, props);
        
        const eksCluster = props.eksCluster;
        const service = props.service;
        const repository = "oci://registry-1.docker.io/bitnamicharts";
        const namespace = props.config.NAMESPACE;

        const generateRandomPassword = (length: number) => {
            return crypto.randomBytes(length).toString('hex').slice(0, length);
        };
        const rabbitMQPassword = generateRandomPassword(12);

        new helm.HelmChart(this, "RedisHelmChart", {
            cluster: eksCluster,
            chart: "redis",
            namespace: service + namespace,
            release: "redis",
            version: "21.2.3",
            wait: false,
            // LOCAVORA https://github.com/aws/aws-cdk/issues/32187 
            repository: repository + "/redis",
            values: {
                auth: {
                    enabled: false
                },
                replica: {
                    replicaCount: 0
                },
                master: {
                    persistence: {
                        storageClass: "gp2"
                    }
                }
            }
        });

        new helm.HelmChart(this, "MongoDBHelmChart", {
            cluster: eksCluster,
            chart: "mongodb",
            namespace: service + namespace,
            release: "mongodb",
            version: "16.5.21",
            wait: false,
            // LOCAVORA https://github.com/aws/aws-cdk/issues/32187
            repository: repository + "/mongodb",
            values: {
                persistence: {
                    storageClass: "gp2"
                }
            }
        });

        new helm.HelmChart(this, "RabbitMQHelmChart", {
            cluster: eksCluster,
            chart: "rabbitmq",
            namespace: service + namespace,
            release: "rabbitmq",
            version: "16.0.7",            
            wait: false,
            // LOCAVORA https://github.com/aws/aws-cdk/issues/32187
            repository: repository + "/rabbitmq",
            values: {
                persistence: {
                    enabled: true,
                    storageClass: "gp2"
                },
                auth: {
                    username: "beckn",
                    password: "beckn1234"
                }
            }
        });

        // new cdk.CfnOutput(this, String("RabbimqPassword"), {
        //     value: rabbitMQPassword,
        // });

    }
}