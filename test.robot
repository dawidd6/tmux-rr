*** Settings ***
Documentation       Test cases for tmux-rr

Library             Dialogs
Library             Process

Suite Setup         Suite Setup
Test Setup          Test Setup
Test Teardown       Test Teardown


*** Variables ***
${CONTAINERFILE}=       Containerfile.test
${CONTAINER}=           tmux-rr-test
${IMAGE}=               tmux-rr-test
${USER}=                test
${UID}=                 1000


*** Test Cases ***
Test
    [Documentation]  TODO
    Run Command In Container    tmux -N new-session -d -s test123
    ${before}=    Run Command In Container
    ...    tmux -N list-panes -a -F '#{session_name}\t#{window_name}\t#{window_layout}'
    Restart Container    tmux=True
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
    ...    podman exec -u ${UID} -e XDG_RUNTIME_DIR=/run/user/1000 ${CONTAINER} ${cmd}
    ...    expected_rc=${expected_rc}
    RETURN    ${result}

Wait For Tmux In Container
    [Documentation]  TODO
    Wait Until Keyword Succeeds    20x    1s    Run Command In Container    tmux -N start-server

Wait For Systemd In Container
    [Documentation]   Block execution until systemd is fully running in the test container
    Wait Until Keyword Succeeds    20x    1s    Run Command In Container    sudo systemctl is-system-running --wait

Start Container
    [Documentation]   Run test container and wait for systemd
    Run Command    podman run -dt --name ${CONTAINER} -v .:/wd -w /wd ${IMAGE}
    Wait For Systemd In Container

Restart Container
    [Documentation]   Restart test container and wait for systemd
    [Arguments]    ${tmux}=False
    Run Command    podman restart ${CONTAINER}
    Wait For Systemd In Container
    IF    ${tmux}    Wait For Tmux In Container

Stop Container
    [Documentation]   Remove test container forcefully
    Run Command    podman rm -f ${CONTAINER}

Install Project
    [Documentation]   Perform project installation steps
    Run Command In Container    sudo make install
    Run Command In Container    systemctl --user enable --now tmux-rr

Test Setup
    [Documentation]  Steps executed before every test case
    Start Container
    Install Project

Test Teardown
    [Documentation]    Steps executed after every test case
    Stop Container

Suite Setup
    [Documentation]    Steps executed before all test cases once
    Log To Console    Container image will be built, please be patient.
    Run Command    podman build -t ${IMAGE} -f ${CONTAINERFILE} .
