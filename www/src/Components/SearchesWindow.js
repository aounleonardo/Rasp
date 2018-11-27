import React, {Component} from 'react';
import {Button, Col} from 'react-bootstrap';

export default class SearchesWindow extends Component {

    render() {
        return (
            <Col md={4}>
                {this.props.searches.map((search) =>
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
    <Button onClick={() => download(
        filename,
        metafile,
        () => console.log("downloading", filename, chunkCount, metafile),
    )}>
        {filename} {chunkCount}
    </Button>;