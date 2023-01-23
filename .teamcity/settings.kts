import jetbrains.buildServer.configs.kotlin.*
import jetbrains.buildServer.configs.kotlin.buildFeatures.dockerSupport
import jetbrains.buildServer.configs.kotlin.buildFeatures.golang
import jetbrains.buildServer.configs.kotlin.buildSteps.ScriptBuildStep
import jetbrains.buildServer.configs.kotlin.buildSteps.dockerCompose
import jetbrains.buildServer.configs.kotlin.buildSteps.script
import jetbrains.buildServer.configs.kotlin.triggers.vcs

/*
The settings script is an entry point for defining a TeamCity
project hierarchy. The script should contain a single call to the
project() function with a Project instance or an init function as
an argument.

VcsRoots, BuildTypes, Templates, and subprojects can be
registered inside the project using the vcsRoot(), buildType(),
template(), and subProject() methods respectively.

To debug settings scripts in command-line, run the

    mvnDebug org.jetbrains.teamcity:teamcity-configs-maven-plugin:generate

command and attach your debugger to the port 8000.

To debug in IntelliJ Idea, open the 'Maven Projects' tool window (View
-> Tool Windows -> Maven Projects), find the generate task node
(Plugins -> teamcity-configs -> teamcity-configs:generate), the
'Debug' option is available in the context menu for the task.
*/

version = "2022.10"

project {

    buildType(TC_TerraformProvider_Test)
}

object TC_TerraformProvider_Test : BuildType({
    id("Test")
    name = "Test"

    params {
        param("env.TF_ACC_PROVIDER_NAMESPACE", "jetbrains")
        param("env.CGO_ENABLED", "0")
        param("env.TF_ACC", "1")
        param("env.GOFLAGS", "-json")
    }

    vcs {
        root(DslContext.settingsRoot)
    }

    steps {
        script {
            name = "Fix permissions"
            scriptContent = "chmod 666 testdata/teamcity.properties"
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
            dockerImage = "curlimages/curl"
            dockerImagePlatform = ScriptBuildStep.ImagePlatform.Linux
        }
        script {
            name = "Run tests"
            scriptContent = "go test ./..."
            dockerImage = "golang:1.19.2-alpine3.16"
            dockerImagePlatform = ScriptBuildStep.ImagePlatform.Linux
        }
    }

    triggers {
        vcs {
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
})
