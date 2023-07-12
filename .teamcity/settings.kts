import jetbrains.buildServer.configs.kotlin.*
import jetbrains.buildServer.configs.kotlin.buildFeatures.dockerSupport
import jetbrains.buildServer.configs.kotlin.buildFeatures.golang
import jetbrains.buildServer.configs.kotlin.buildFeatures.sshAgent
import jetbrains.buildServer.configs.kotlin.buildSteps.ScriptBuildStep
import jetbrains.buildServer.configs.kotlin.buildSteps.dockerCompose
import jetbrains.buildServer.configs.kotlin.buildSteps.script
import jetbrains.buildServer.configs.kotlin.triggers.vcs

version = "2023.05"

project {
    buildType(Test)
    buildType(Release)
    buildTypesOrder = listOf(Test, Release)
}

object Test : BuildType({
    id("Test")
    name = "Test"

    vcs {
        root(DslContext.settingsRoot)
    }

    dependencies {
        dependency(AbsoluteId("TC_Trunk_DistParts_PluginRestApi")) {
            artifacts {
                buildRule = lastSuccessful("+:mkuzmin/auth")
                artifactRules = "rest-api.zip => testdata/plugins/"
            }
        }
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
            forcePull = true
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
            dockerImage = "registry.jetbrains.team/p/tc/docker/teamcity-server-staging:2023.05.1-linux"
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

object Release : BuildType({
    id("Release")
    name = "Build & Release"

    type = Type.DEPLOYMENT
    enablePersonalBuilds = false
    maxRunningBuilds = 1

    dependencies {
        snapshot(Test) {
            onDependencyFailure = FailureAction.FAIL_TO_START
        }
    }

    vcs {
        root(DslContext.settingsRoot)
        branchFilter = "+:<default>"
        checkoutMode = CheckoutMode.ON_AGENT
    }

    params {
        password(name = "env.GPG_PRIVATE_KEY", value = "credentialsJSON:ab4c79dc-954c-481f-8bd5-23ef5f18f8a2")
        password(name = "env.GITHUB_TOKEN", value = "credentialsJSON:449b199b-427f-49e1-95d1-8254b938f0b5")
    }

    features {
        sshAgent {
            teamcitySshKey = "github-mkuzmin"
        }
    }

    steps {
        script {
            name = "Tag current git commit"
            scriptContent = """
                git tag "v0.0.%build.number%"
                git push origin "v0.0.%build.number%"
            """.trimIndent()
        }
        script {
            name = "Build & release"
            dockerImage = "goreleaser/goreleaser:v1.18.2"
            dockerImagePlatform = ScriptBuildStep.ImagePlatform.Linux
            scriptContent = """
                git config --global --add safe.directory %teamcity.build.checkoutDir%
                ./release.sh
            """.trimIndent()
        }
    }
})
