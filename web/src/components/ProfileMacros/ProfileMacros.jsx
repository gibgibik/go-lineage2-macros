import {Macros} from "../Macros/Macros.jsx";
import React from "react";
import {Box, Checkbox, FormControl, FormControlLabel, FormGroup} from "@mui/material";

export const ProfileMacros = ({isBatchRun, profileName, profiles, setProfiles, ...props}) => {
    const {presetId} = props;
    const isActiveChangeHandler = (event) => {
        setProfiles({
            ...profiles,
            [profileName]: {
                ...profiles[profileName],
                items: [...(profiles[profileName].items || []).map((item) => {
                    if (item.preset.id == presetId) {
                        item.batch_run = event.target.checked;
                    }
                    return item;
                })],
            }
        });
    }
    return <Box>
        <FormGroup sx={{m: 2}}>
            <FormControlLabel control={<Checkbox defaultChecked onChange={isActiveChangeHandler} checked={isBatchRun}/>} label="Batch run(based on first item conditions)" />
        </FormGroup>
        <Macros {...props} />
    </Box>;
}
