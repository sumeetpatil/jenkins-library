import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/codeqlExecuteScan.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    for ( e in script.commonPipelineEnvironment ) {
            print "key = ${e.key}, value = ${e.value}"
    }
    print "token = ${env.PIPER_githubToken}"

    List credentials = [[type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
