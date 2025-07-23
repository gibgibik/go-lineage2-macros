import {useEffect, useState} from 'react';
import {QueryBuilder} from 'react-querybuilder';
import {QueryBuilderMaterial} from "@react-querybuilder/material";
import {Button, FormControl} from "@mui/material";

const fields = [
    {name: 'target_hp', label: 'Target HP'},
    {name: 'my_hp', label: 'My HP'},
    {name: 'my_mp', label: 'My MP'},
    {name: 'since_last_success_target', label: 'Since Last Success Target'},
    {name: 'full_target_hp_unchanged_since', label: 'Full Target Hp Unchanged'},
    {name: 'party_member_hp_1', label: 'Party Member 1 HP'},
    {name: 'party_member_hp_2', label: 'Party Member 2 HP'},
    {name: 'party_member_hp_3', label: 'Party Member 3 HP'},
    {name: 'party_member_hp_4', label: 'Party Member 4 HP'},
    {name: 'party_member_hp_5', label: 'Party Member 5 HP'},
    {name: 'party_member_hp_6', label: 'Party Member 6 HP'},
    {name: 'party_member_hp_7', label: 'Party Member 7 HP'},
    {name: 'party_member_hp_8', label: 'Party Member 8 HP'}
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
    {name: 'OR', label: 'OR'},
];
const muiComponents = {
    Button: (props) => <Button onClick={props.onClick}>{props.className === 'rule-remove' ? '-' : '+'}</Button>
};
export const Condition = ({onQueryChange, fullWidth, conditions}) => {
    const [query, setQuery] = useState(conditions);
    useEffect(() => {
        setQuery(conditions);
    }, [conditions]);
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
