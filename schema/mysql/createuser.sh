#!/bin/bash
echo "CREATE USER '${1:?}'@'%' IDENTIFIED  BY '${2:?}';  GRANT SELECT,INSERT,UPDATE ON passwdpot.* to  '${1}'@'%' ;"
