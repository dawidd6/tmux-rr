*** Settings ***
Documentation       Test cases for tmux-rr

Library             Process

Test Setup          Test Setup
Test Teardown       Test Teardown


*** Variables ***
${CONTAINERFILE}=       Containerfile.test
${CONTAINER}=           tmux-rr-test
${IMAGE}=               tmux-rr-test
${USER}=                test
${UID}=                 1000
${GID}=                 1000


*** Test Cases ***
Test
    [Documentation]  TODO
    Run Command In Container    tmux -N new-session -d -s test123
    ${before}=    Run Command In Container
    ...    tmux -N list-panes -a -F '#{session_name}\t#{window_name}\t#{window_layout}'
    ${after}=    Run Command In Container
    ...    tmux -N list-panes -a -F '#{session_name}\t#{window_name}\t#{window_layout}'
    Should Be Equal As Strings    ${before}    ${after}


*** Keywords ***
Run Command
    [Documentation]   Execute a command on the host, check exit code and return result
    [Arguments]    ${cmd}    ${expected_rc}=0
    ${result}=    Run Process    ${cmd}    shell=True
    Log Many    ${result.rc}    ${result.stdout}    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    ${{int($expected_rc)}}
    RETURN    ${result}

Run Command In Container
    [Documentation]   Execute a command in the test container, check exit code and return result
    [Arguments]    ${cmd}    ${expected_rc}=0
    ${result}=    Run Command
    ...    podman exec -u ${UID}:${GID} -e XDG_RUNTIME_DIR=/run/user/${UID} ${CONTAINER} ${cmd}
    ...    expected_rc=${expected_rc}
    RETURN    ${result}

Wait For Tmux In Container
    [Documentation]  TODO
    Wait Until Keyword Succeeds    20x    1s    Run Command In Container    tmux -N start-server

Wait For Systemd In Container
    [Documentation]   Block execution until systemd is fully running in the test container
    Wait Until Keyword Succeeds    20x    1s    Run Command In Container    systemctl is-system-running --wait --user

Run Container
    [Documentation]   Run test container and wait for systemd
    Run Command    podman run -dt --name ${CONTAINER} -v .:/wd -w /wd ${IMAGE}
    Wait For Systemd In Container

Restart Container
    [Documentation]   Restart test container and wait for systemd
    Run Command    podman restart ${CONTAINER}
    Wait For Systemd In Container

Remove Container
    [Documentation]   Remove test container forcefully
    Run Command    podman rm -f ${CONTAINER}

Test Setup
    [Documentation]  Steps executed before every test case
    Run Container

Test Teardown
    [Documentation]    Steps executed after every test case
    Remove Container
