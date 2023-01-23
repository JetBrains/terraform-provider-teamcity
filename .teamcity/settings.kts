import jetbrains.buildServer.configs.kotlin.*
import jetbrains.buildServer.configs.kotlin.buildFeatures.dockerSupport
import jetbrains.buildServer.configs.kotlin.buildFeatures.golang
import jetbrains.buildServer.configs.kotlin.buildSteps.ScriptBuildStep
import jetbrains.buildServer.configs.kotlin.buildSteps.dockerCompose
import jetbrains.buildServer.configs.kotlin.buildSteps.script
import jetbrains.buildServer.configs.kotlin.triggers.vcs

version = "2022.10"

project {
    buildType(TC_TerraformProvider_Test)
}

object TC_TerraformProvider_Test : BuildType({
    id("Test")
    name = "Test"

    vcs {
        root(DslContext.settingsRoot)
    }

    features {
        dockerSupport {
            loginToRegistry = on {
                dockerRegistryId = "PROJECT_EXT_789,PROJECT_EXT_315"
            }
        }
        golang {
            testFormat = "json"
        }
    }

    steps {
        script {
            name = "Fix permissions"
            scriptContent = "chmod 666 testdata/teamcity.properties" // TeamCity requires a write lock
        }

        dockerCompose {
            name = "Start TeamCity"
            file = "docker-compose.yml"
        }

        script {
            name = "Get token"
            scriptContent = """
                set -e
                json=${'$'}(curl -X POST http://teamcity-server:8111/app/rest/users/current/tokens/test -H "Authorization: Basic OnRva2VuMTIz" -H "Accept: application/json")
                echo ${'$'}json
                token=${'$'}(echo ${'$'}json | sed -n 's|.*"value":"\([^"]*\)".*|\1|p')
                echo ${'$'}token
                echo "##teamcity[setParameter name='env.TEAMCITY_TOKEN' value='${'$'}token']"
                echo "##teamcity[setParameter name='env.TEAMCITY_HOST' value='http://teamcity-server:8111']"
            """.trimIndent()
            dockerImage = "registry.jetbrains.team/p/tc/docker/teamcity-server-staging:EAP-linux"
            dockerImagePlatform = ScriptBuildStep.ImagePlatform.Linux
        }

        script {
            name = "Run tests"
            scriptContent = """
                export CGO_ENABLED=0
                export TF_ACC=1
                export TF_ACC_PROVIDER_NAMESPACE=jetbrains
                go test -json ./...
            """.trimIndent()
            dockerImage = "golang:1.19.2-alpine3.16"
            dockerImagePlatform = ScriptBuildStep.ImagePlatform.Linux
        }
    }

    triggers {
        vcs {
        }
    }
})
