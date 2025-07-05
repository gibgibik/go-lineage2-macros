import {FormControl, InputLabel, MenuItem, Select} from "@mui/material";
import {useState} from "react";

const rawActions = [
    '/assist',
    '/attack',
    '/target',
    '/delay',
    '/useskill',
    '/press',
];
export const MacrosAction = ({name, initValue}) => {
    const [value, setValue] = useState(null);
    return <FormControl fullWidth={true}>
        <InputLabel id="action-select-label" >Action</InputLabel>
        <Select
            variant={'outlined'}
            id="action-simple-select"
            label="Action"
            name={name}
            value={value === null ? initValue : value}
            onChange={(e) => setValue(e.target.value)}
        >
            <MenuItem value={""} />
            {rawActions.map((item, k) => <MenuItem key={k} value={item}>{item}</MenuItem>)}
        </Select>
    </FormControl>;
}
