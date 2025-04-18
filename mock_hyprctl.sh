#!/bin/bash

if [[ "$1" != hyprsunset ]]
then
    echo "unknown request"
    exit 0
fi

if [[ -z "$2" ]]
then
    echo "invalid command"
    exit 0
fi

# hyprsunset currently always returns 0

case "$2" in
    temperature)
        # No argument given will print current temperature
        if [[ -z "$3" ]]
        then
            echo 6500
        else
            # Argument supplied, return ok
            # Any non-integer argument here will make hyprsunset crash.
            # No need to test :)
            echo ok
        fi
        ;;
    gamma)
        # No argument given will print current gamma
        if [[ -z "$3" ]]
        then
            echo 100
        else
            # Argument supplied, check range
            if [[ "$3" -ge 0 ]] && [[ "$3" -le 100 ]]
            then
                echo ok
            else
                echo "Invalid gamma value (should be in range 0-100%)"
            fi
        fi
        ;;
    *)
        echo "invalid command"
        ;;
esac

exit 0
