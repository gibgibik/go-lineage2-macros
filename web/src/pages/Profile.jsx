import React from "react";
import {Box} from "@mui/material";

export const Profile = ({value, index, ...other}) => {
    if (value !== index) {
        return null;
    }
    return (
        <div
            role="tabpanel"
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            {value === index && <Box sx={{p: 3}}>asd</Box>}
        </div>
    );
}
