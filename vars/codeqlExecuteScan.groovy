import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/codeqlExecuteScan.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    print "commitId = ${script.commonPipelineEnvironment.gitCommitId}"

    List credentials = [[type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
