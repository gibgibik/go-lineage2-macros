import {FormControl, InputLabel, MenuItem, Select} from "@mui/material";
import {useState} from "react";

const rawActions = [
    '/assistpartymember',
    '/assist',
    '/attack',
    '/target',
    '/targetnext',
    '/delay',
    '/press',
    '/pickup',
    '/aitargetnext',
    '/stop',
];
export const MacrosAction = ({name, initValue}) => {
    const [value, setValue] = useState(null);
    return <FormControl sx={{width: '250px'}}>
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
