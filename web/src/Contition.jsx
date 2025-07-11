import {useEffect, useState} from 'react';
import {QueryBuilder} from 'react-querybuilder';
import {QueryBuilderMaterial} from "@react-querybuilder/material";
import {Button, FormControl} from "@mui/material";

const fields = [
    {name: 'target_hp', label: 'Target HP'},
    {name: 'my_hp', label: 'My HP'},
    {name: 'my_mp', label: 'My MP'},
    {name: 'since_last_success_target', label: 'Since Last Success Target'},
];

const controlElements = {
    addGroupAction: () => null,
}

const operators = [
    {name: '>', label: '>'},
    {name: '=', label: '='},
    {name: '<', label: '<'},
];
const combinators = [
    {name: 'AND', label: 'AND'},
];
const muiComponents = {
    Button: (props) => <Button onClick={props.onClick}>{props.className === 'rule-remove' ? '-' : '+'}</Button>
};
export const Condition = ({onQueryChange, fullWidth, conditions}) => {
    const [query, setQuery] = useState(conditions);
    useEffect(() => {
        onQueryChange(query.rules, 'json');
    }, [query]);
    return (
        <FormControl fullWidth={fullWidth}>
            <QueryBuilderMaterial muiComponents={muiComponents}>
                <QueryBuilder fields={fields} query={query} onQueryChange={setQuery} controlElements={controlElements}
                              operators={operators} combinators={combinators}/>
            </QueryBuilderMaterial>
        </FormControl>
    );
}
