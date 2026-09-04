*** Settings ***
Documentation    Test cases for tmux-rr

Library          Process

Test Setup       Test Setup
Test Teardown    Test Teardown


*** Variables ***
${CONTAINER}=    tmux-rr
${IMAGE}=        tmux-rr
${UID}=          1000
${GID}=          1000


*** Test Cases ***
Fish
    [Documentation]  TODO
    Run Command In Container    tmux -N new-session -d -s test123   expected_stdout=False
    ${before}=    Run Command In Container
    ...    tmux -N list-panes -a -F '#{session_name} #{window_name} #{window_layout}'
    ${after}=    Run Command In Container
    ...    tmux -N list-panes -a -F '#{session_name} #{window_name} #{window_layout}'
    Should Be Equal As Strings    ${before}    ${after}


*** Keywords ***
Run Command
    [Documentation]   Execute a command on the host, check exit code and return result
    [Arguments]    ${cmd}    ${expected_rc}=0    ${expected_stdout}=True    ${expected_stderr}=False
    ${result}=    Run Process    ${cmd}    shell=True
    Log Many    ${result.rc}    ${result.stdout}    ${result.stderr}
    IF    ${expected_stdout}
        Should Not Be Empty    ${result.stdout}
    ELSE
        Should Be Empty    ${result.stdout}
    END
    IF    ${expected_stderr}
        Should Not Be Empty    ${result.stderr}
    ELSE
        Should Be Empty    ${result.stderr}
    END
    Should Be Equal As Integers    ${result.rc}    ${{int($expected_rc)}}
    RETURN    ${result}

Run Command In Container
    [Documentation]   Execute a command in the test container, check exit code and return result
    [Arguments]    ${cmd}    ${expected_rc}=0    ${expected_stdout}=True    ${expected_stderr}=False
    ${result}=    Run Command
    ...    podman exec -u ${UID}:${GID} -e XDG_RUNTIME_DIR=/run/user/${UID} ${CONTAINER} ${cmd}
    ...    expected_rc=${expected_rc}    expected_stdout=${expected_stdout}    expected_stderr=${expected_stderr}
    RETURN    ${result}

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
