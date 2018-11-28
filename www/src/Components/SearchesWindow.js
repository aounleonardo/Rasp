import React, {Component} from 'react';
import {Button, Col} from 'react-bootstrap';

export default class SearchesWindow extends Component {

    render() {
        return (
            <Col md={10}>
                {this.props.searches
                    .filter((search) => this.props.keywords.some(
                        (word) => search.Filename.includes(word)),
                    )
                    .map((search) =>
                        <SearchResult
                            key={search.Metakey}
                            filename={search.Filename}
                            chunkCount={search.ChunkCount}
                            metafile={search.Metakey}
                            download={this.props.download}
                        />
                    )}
            </Col>
        )
    }
}

const SearchResult = ({filename, chunkCount, metafile, download}) =>
    <Button
        onClick={() => download(
            filename,
            metafile,
            () => console.log("downloading", filename, chunkCount, metafile),
        )}
        style={{
            height: '25%',
            width: '25%',
            color: 'dodgerblue',
            fontSize: '110%',
            fontWeight: 'bold',
        }}
    >
        {filename} {chunkCount}
    </Button>;