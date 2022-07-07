import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/checkmarxExecuteScan.yaml'

//Metadata maintained in file project://resources/metadata/checkmarxExecuteScan.yaml

void call(Map parameters = [:]) {
    echo "checkmarx1"
    List credentials = [[type: 'usernamePassword', id: 'checkmarxCredentialsId', env: ['PIPER_username', 'PIPER_password']]]
    echo "checkmarx2"
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, true)
}
